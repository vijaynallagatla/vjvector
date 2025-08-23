package cluster

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config represents cluster configuration
type Config struct {
	NodeAddress string `json:"node_address"`
	NodePort    int    `json:"node_port"`
	ShardCount  int    `json:"shard_count"`
}

// EtcdCluster implements the Cluster interface using etcd for coordination
type EtcdCluster struct {
	config             *Config
	etcdClient         *clientv3.Client
	nodeInfo           *NodeInfo
	role               NodeRole
	master             *NodeInfo
	peers              map[string]*Peer
	mu                 sync.RWMutex
	logger             *slog.Logger
	ctx                context.Context
	cancel             context.CancelFunc
	peerManager        PeerManager
	replicationManager ReplicationManager
	healthChecker      HealthChecker
	metricsCollector   MetricsCollector
	consensus          ConsensusProtocol
	sharding           ShardingStrategy
	loadBalancer       LoadBalancer
}

// NewEtcdCluster creates a new etcd-based cluster
func NewEtcdCluster(config *Config, etcdClient *clientv3.Client, logger *slog.Logger) (*EtcdCluster, error) {
	nodeID, err := generateNodeID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate node ID: %w", err)
	}

	nodeInfo := &NodeInfo{
		ID:        nodeID,
		Address:   config.NodeAddress,
		Port:      config.NodePort,
		Role:      NodeRoleSlave, // Start as slave, will be promoted if needed
		State:     NodeStateStarting,
		StartTime: time.Now(),
	}

	cluster := &EtcdCluster{
		config:     config,
		etcdClient: etcdClient,
		nodeInfo:   nodeInfo,
		role:       NodeRoleSlave,
		peers:      make(map[string]*Peer),
		logger:     logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cluster.ctx = ctx
	cluster.cancel = cancel

	if err := cluster.initializeManagers(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize managers: %w", err)
	}

	return cluster, nil
}

// initializeManagers initializes all cluster managers
func (c *EtcdCluster) initializeManagers() error {
	// Initialize peer manager
	c.peerManager = NewEtcdPeerManager(c.etcdClient, c.logger)

	// Initialize replication manager
	c.replicationManager = NewEtcdReplicationManager(c.etcdClient, c.logger)

	// Initialize health checker
	c.healthChecker = NewEtcdHealthChecker(c.etcdClient, c.logger)

	// Initialize metrics collector
	c.metricsCollector = NewEtcdMetricsCollector(c.etcdClient, c.logger)

	// Initialize consensus protocol
	c.consensus = NewEtcdConsensusProtocol(c.etcdClient, c.logger)

	// Initialize sharding strategy
	c.sharding = NewHashSharding(c.config.ShardCount)

	// Initialize load balancer
	c.loadBalancer = NewRoundRobinLoadBalancer()

	return nil
}

// Start starts the cluster
func (c *EtcdCluster) Start(ctx context.Context) error {
	c.logger.Info("Starting etcd cluster", "node_id", c.nodeInfo.ID, "address", c.nodeInfo.Address)

	// Register node with etcd
	if err := c.registerNode(ctx); err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}

	// Start heartbeat
	go c.startHeartbeat()

	// Watch for cluster changes
	go c.watchClusterChanges()

	// Discover existing nodes
	if err := c.discoverNodes(ctx); err != nil {
		c.logger.Warn("Failed to discover existing nodes", "error", err)
	}

	// Try to become master if no master exists
	go c.tryBecomeMaster()

	c.nodeInfo.State = NodeStateRunning
	c.logger.Info("✅ Etcd cluster started successfully", "node_id", c.nodeInfo.ID, "role", c.role)

	return nil
}

// Stop stops the cluster
func (c *EtcdCluster) Stop(ctx context.Context) error {
	c.logger.Info("🛑 Stopping etcd cluster", "node_id", c.nodeInfo.ID)

	c.cancel()

	// Deregister node from etcd
	if err := c.deregisterNode(ctx); err != nil {
		c.logger.Error("Failed to deregister node", "error", err)
	}

	c.nodeInfo.State = NodeStateStopped
	c.logger.Info("✅ Etcd cluster stopped", "node_id", c.nodeInfo.ID)

	return nil
}

// GetNodeInfo returns the current node information
func (c *EtcdCluster) GetNodeInfo() *NodeInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodeInfo
}

// GetRole returns the current node role
func (c *EtcdCluster) GetRole() NodeRole {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.role
}

// GetMaster returns the current master node
func (c *EtcdCluster) GetMaster() *NodeInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.master
}

// Health returns the health status of the cluster
func (c *EtcdCluster) Health(ctx context.Context) (*ClusterHealth, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Get cluster health from health checker
	if c.healthChecker != nil {
		return c.healthChecker.CheckHealth(ctx)
	}

	// Fallback to basic health check
	status := "healthy"
	if c.nodeInfo.State != NodeStateRunning {
		status = "unhealthy"
	}

	return &ClusterHealth{
		Status:      status,
		NodeCount:   len(c.peers) + 1, // +1 for self
		MasterID:    c.master.ID,
		Replicas:    len(c.peers),
		Shards:      c.config.ShardCount,
		LastUpdated: time.Now(),
		Details: map[string]interface{}{
			"node_state": c.nodeInfo.State,
			"role":       c.role,
		},
	}, nil
}

// IsMaster returns true if this node is the master
func (c *EtcdCluster) IsMaster() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.role == NodeRoleMaster
}

// GetPeers returns all peer nodes
func (c *EtcdCluster) GetPeers() []*Peer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	peers := make([]*Peer, 0, len(c.peers))
	for _, peer := range c.peers {
		peers = append(peers, peer)
	}
	return peers
}

// GetHealth returns the cluster health
func (c *EtcdCluster) GetHealth() *ClusterHealth {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := c.healthChecker.CheckHealth(ctx)
	if err != nil {
		return &ClusterHealth{
			Status:      "unhealthy",
			NodeCount:   0,
			MasterID:    "",
			Replicas:    0,
			Shards:      0,
			LastUpdated: time.Now(),
			Details:     map[string]interface{}{"error": err.Error()},
		}
	}

	return health
}

// GetMetrics returns the cluster metrics
func (c *EtcdCluster) GetMetrics() *ClusterMetrics {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metrics, err := c.metricsCollector.CollectMetrics(ctx)
	if err != nil {
		c.logger.Error("Failed to collect metrics", "error", err)
		return &ClusterMetrics{
			Timestamp: time.Now(),
			Details:   map[string]interface{}{"error": err.Error()},
		}
	}

	return metrics
}

// registerNode registers the node with etcd
func (c *EtcdCluster) registerNode(ctx context.Context) error {
	nodeData, err := json.Marshal(c.nodeInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal node info: %w", err)
	}

	key := fmt.Sprintf("/vjvector/nodes/%s", c.nodeInfo.ID)
	_, err = c.etcdClient.Put(ctx, key, string(nodeData))
	if err != nil {
		return fmt.Errorf("failed to register node with etcd: %w", err)
	}

	c.logger.Info("Node registered with etcd", "node_id", c.nodeInfo.ID, "key", key)
	return nil
}

// deregisterNode removes the node from etcd
func (c *EtcdCluster) deregisterNode(ctx context.Context) error {
	key := fmt.Sprintf("/vjvector/nodes/%s", c.nodeInfo.ID)
	_, err := c.etcdClient.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to deregister node from etcd: %w", err)
	}

	c.logger.Info("Node deregistered from etcd", "node_id", c.nodeInfo.ID, "key", key)
	return nil
}

// startHeartbeat starts the heartbeat mechanism
func (c *EtcdCluster) startHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.sendHeartbeat(); err != nil {
				c.logger.Error("Failed to send heartbeat", "error", err)
			}
		}
	}
}

// sendHeartbeat sends a heartbeat to etcd
func (c *EtcdCluster) sendHeartbeat() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update node info with current timestamp
	c.nodeInfo.LastSeen = time.Now()

	nodeData, err := json.Marshal(c.nodeInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal node info: %w", err)
	}

	key := fmt.Sprintf("/vjvector/nodes/%s", c.nodeInfo.ID)
	_, err = c.etcdClient.Put(ctx, key, string(nodeData))
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}

	return nil
}

// watchClusterChanges watches for changes in the cluster
func (c *EtcdCluster) watchClusterChanges() {
	watchChan := c.etcdClient.Watch(c.ctx, "/vjvector/nodes/", clientv3.WithPrefix())

	for {
		select {
		case <-c.ctx.Done():
			return
		case watchResp := <-watchChan:
			for _, ev := range watchResp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					c.handleNodeJoin(string(ev.Kv.Key), ev.Kv.Value)
				case clientv3.EventTypeDelete:
					key := string(ev.Kv.Key)
					if len(key) > 13 {
						nodeID := key[13:] // Remove "/vjvector/nodes/" prefix
						c.handleNodeLeave(nodeID)
					}
				}
			}
		}
	}
}

// tryBecomeMaster tries to become the master node
func (c *EtcdCluster) tryBecomeMaster() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if there's already a master
	resp, err := c.etcdClient.Get(ctx, "/vjvector/master", clientv3.WithPrefix())
	if err != nil {
		c.logger.Error("Failed to check for existing master", "error", err)
		return
	}

	if len(resp.Kvs) == 0 {
		// No master exists, try to become one
		if err := c.becomeMaster(ctx); err != nil {
			c.logger.Error("Failed to become master", "error", err)
		}
	}
}

// becomeMaster makes this node the master
func (c *EtcdCluster) becomeMaster(ctx context.Context) error {
	masterData := map[string]interface{}{
		"node_id":   c.nodeInfo.ID,
		"address":   c.nodeInfo.Address,
		"port":      c.nodeInfo.Port,
		"timestamp": time.Now(),
	}

	masterBytes, err := json.Marshal(masterData)
	if err != nil {
		return fmt.Errorf("failed to marshal master data: %w", err)
	}

	// Use a transaction to ensure atomicity
	txn := c.etcdClient.Txn(ctx)
	txn.If(clientv3.Compare(clientv3.CreateRevision("/vjvector/master"), "=", 0))
	txn.Then(clientv3.OpPut("/vjvector/master", string(masterBytes)))

	resp, err := txn.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit master transaction: %w", err)
	}

	if resp.Succeeded {
		c.mu.Lock()
		c.role = NodeRoleMaster
		c.nodeInfo.Role = NodeRoleMaster
		c.master = c.nodeInfo
		c.mu.Unlock()

		c.logger.Info("🎉 Node became master", "node_id", c.nodeInfo.ID)
		return nil
	}

	c.logger.Info("Another node became master", "node_id", c.nodeInfo.ID)
	return nil
}

// handleNodeJoin handles node join events
func (c *EtcdCluster) handleNodeJoin(key string, value []byte) {
	nodeID := key[13:] // Remove "/vjvector/nodes/" prefix

	if nodeID == c.nodeInfo.ID {
		return // Ignore our own join event
	}

	var nodeInfo NodeInfo
	if err := json.Unmarshal(value, &nodeInfo); err != nil {
		c.logger.Error("Failed to unmarshal node info", "error", err)
		return
	}

	c.logger.Info("Node joined", "node_id", nodeID, "address", nodeInfo.Address)

	// TODO: Establish connection to the new node
}

// handleNodeLeave handles node leave events
func (c *EtcdCluster) handleNodeLeave(nodeID string) {
	if nodeID == c.nodeInfo.ID {
		return // Ignore our own leave event
	}

	c.logger.Info("Node left", "node_id", nodeID)

	// Remove peer
	c.mu.Lock()
	if peer, exists := c.peers[nodeID]; exists {
		if peer.Connection != nil {
			if err := peer.Connection.Close(); err != nil {
				c.logger.Error("Failed to close peer connection", "error", err)
			}
		}
		delete(c.peers, nodeID)
	}
	c.mu.Unlock()
}

// discoverNodes discovers existing nodes in the cluster
func (c *EtcdCluster) discoverNodes(ctx context.Context) error {
	resp, err := c.etcdClient.Get(ctx, "/vjvector/nodes/", clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to get nodes from etcd: %w", err)
	}

	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		if len(key) <= 13 {
			continue
		}

		nodeID := key[13:] // Remove "/vjvector/nodes/" prefix
		if nodeID == c.nodeInfo.ID {
			continue // Skip ourselves
		}

		var nodeInfo NodeInfo
		if err := json.Unmarshal(kv.Value, &nodeInfo); err != nil {
			c.logger.Error("Failed to unmarshal node info", "error", err)
			continue
		}

		c.logger.Info("Discovered existing node", "node_id", nodeID, "address", nodeInfo.Address)

		// TODO: Establish connection to the discovered node
	}

	return nil
}

// generateNodeID generates a random node ID
func generateNodeID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("node-%x", bytes), nil
}

// NewEtcdPeerManager creates a new etcd-based peer manager
func NewEtcdPeerManager(client *clientv3.Client, logger *slog.Logger) PeerManager {
	return &EtcdPeerManager{
		client: client,
		logger: logger,
		peers:  make(map[string]*Peer),
	}
}

// NewEtcdReplicationManager creates a new etcd-based replication manager
func NewEtcdReplicationManager(client *clientv3.Client, logger *slog.Logger) ReplicationManager {
	return &EtcdReplicationManager{
		client: client,
		logger: logger,
	}
}

// NewEtcdHealthChecker creates a new etcd-based health checker
func NewEtcdHealthChecker(client *clientv3.Client, logger *slog.Logger) HealthChecker {
	return &EtcdHealthChecker{
		client: client,
		logger: logger,
	}
}

// NewEtcdMetricsCollector creates a new etcd-based metrics collector
func NewEtcdMetricsCollector(client *clientv3.Client, logger *slog.Logger) MetricsCollector {
	return &EtcdMetricsCollector{
		client: client,
		logger: logger,
	}
}

// NewEtcdConsensusProtocol creates a new etcd-based consensus protocol
func NewEtcdConsensusProtocol(client *clientv3.Client, logger *slog.Logger) ConsensusProtocol {
	return &EtcdConsensusProtocol{
		client: client,
		logger: logger,
	}
}

// EtcdPeerManager implements PeerManager using etcd
type EtcdPeerManager struct {
	client *clientv3.Client
	logger *slog.Logger
	peers  map[string]*Peer
	mu     sync.RWMutex
}

// AddPeer adds a new peer to the manager
func (pm *EtcdPeerManager) AddPeer(peer *Peer) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.peers[peer.Info.ID] = peer
	pm.logger.Info("Peer added", "peer_id", peer.Info.ID, "address", peer.Info.Address)
	return nil
}

// RemovePeer removes a peer from the manager
func (pm *EtcdPeerManager) RemovePeer(peerID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if peer, exists := pm.peers[peerID]; exists {
		if peer.Connection != nil {
			if err := peer.Connection.Close(); err != nil {
				pm.logger.Error("Failed to close peer connection", "error", err)
			}
		}
		delete(pm.peers, peerID)
		pm.logger.Info("Peer removed", "peer_id", peerID)
	}
	return nil
}

// GetPeer retrieves a peer by ID
func (pm *EtcdPeerManager) GetPeer(peerID string) (*Peer, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peer, exists := pm.peers[peerID]
	if !exists {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}
	return peer, nil
}

// GetPeers returns all managed peers
func (pm *EtcdPeerManager) GetPeers() []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peers := make([]*Peer, 0, len(pm.peers))
	for _, peer := range pm.peers {
		peers = append(peers, peer)
	}
	return peers
}

// Broadcast broadcasts a message to all peers
func (pm *EtcdPeerManager) Broadcast(ctx context.Context, message *Message) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, peer := range pm.peers {
		if err := pm.SendToPeer(ctx, peer.Info.ID, message); err != nil {
			pm.logger.Error("Failed to broadcast to peer", "peer_id", peer.Info.ID, "error", err)
		}
	}
	return nil
}

// SendToPeer sends a message to a specific peer
func (pm *EtcdPeerManager) SendToPeer(ctx context.Context, peerID string, message *Message) error {
	_, err := pm.GetPeer(peerID)
	if err != nil {
		return err
	}

	// TODO: Implement actual message sending logic
	pm.logger.Info("Sending message to peer", "peer_id", peerID, "message_type", message.Type)
	return nil
}

// EtcdReplicationManager implements ReplicationManager using etcd
type EtcdReplicationManager struct {
	client *clientv3.Client
	logger *slog.Logger
}

// StartReplication starts replication for a shard
func (rm *EtcdReplicationManager) StartReplication(ctx context.Context, shardID int) error {
	rm.logger.Info("Starting replication", "shard_id", shardID)
	// TODO: Implement actual replication logic
	return nil
}

// StopReplication stops replication for a shard
func (rm *EtcdReplicationManager) StopReplication(ctx context.Context, shardID int) error {
	rm.logger.Info("Stopping replication", "shard_id", shardID)
	// TODO: Implement actual replication stop logic
	return nil
}

// GetReplicationStatus returns the replication status
func (rm *EtcdReplicationManager) GetReplicationStatus(shardID int) (*ReplicationStatus, error) {
	// TODO: Implement actual status retrieval
	return &ReplicationStatus{
		ShardID:      shardID,
		Status:       "unknown",
		LastSync:     time.Now(),
		SyncLag:      0,
		ReplicaCount: 0,
		ReplicaNodes: []string{},
	}, nil
}

// SyncData synchronizes data with peers
func (rm *EtcdReplicationManager) SyncData(ctx context.Context, shardID int) error {
	rm.logger.Info("Syncing data", "shard_id", shardID)
	// TODO: Implement actual data sync logic
	return nil
}

// GetReplicaNodes returns nodes that have replicas of a shard
func (rm *EtcdReplicationManager) GetReplicaNodes(shardID int) []*NodeInfo {
	// TODO: Implement actual replica node retrieval
	return []*NodeInfo{}
}

// EtcdHealthChecker implements HealthChecker using etcd
type EtcdHealthChecker struct {
	client *clientv3.Client
	logger *slog.Logger
}

// CheckHealth checks the health of the cluster
func (hc *EtcdHealthChecker) CheckHealth(ctx context.Context) (*ClusterHealth, error) {
	// TODO: Implement actual cluster health check logic
	return &ClusterHealth{
		Status:      "healthy",
		NodeCount:   1,
		MasterID:    "local",
		Replicas:    0,
		Shards:      0,
		LastUpdated: time.Now(),
	}, nil
}

// CheckNodeHealth checks the health of a specific node
func (hc *EtcdHealthChecker) CheckNodeHealth(ctx context.Context, nodeID string) (*NodeHealth, error) {
	// TODO: Implement actual node health check logic
	return &NodeHealth{
		NodeID:       nodeID,
		Status:       "healthy",
		LastCheck:    time.Now(),
		ResponseTime: 0,
		Details:      map[string]interface{}{"etcd": "connected"},
	}, nil
}

// RegisterHealthCheck registers a custom health check
func (hc *EtcdHealthChecker) RegisterHealthCheck(name string, check HealthCheckFunc) error {
	// TODO: Implement actual health check registration
	hc.logger.Info("Health check registered", "name", name)
	return nil
}

// EtcdMetricsCollector implements MetricsCollector using etcd
type EtcdMetricsCollector struct {
	client *clientv3.Client
	logger *slog.Logger
}

// CollectMetrics collects metrics from the cluster
func (mc *EtcdMetricsCollector) CollectMetrics(ctx context.Context) (*ClusterMetrics, error) {
	// TODO: Implement actual metrics collection
	return &ClusterMetrics{
		Timestamp:      time.Now(),
		NodeCount:      1,
		ActiveNodes:    1,
		TotalRequests:  0,
		ActiveRequests: 0,
		AvgLatency:     0,
		ErrorRate:      0,
		ShardCount:     0,
		ReplicaCount:   0,
	}, nil
}

// GetNodeMetrics returns metrics for a specific node
func (mc *EtcdMetricsCollector) GetNodeMetrics(ctx context.Context, nodeID string) (*NodeMetrics, error) {
	// TODO: Implement actual node metrics collection
	return &NodeMetrics{
		NodeID:       nodeID,
		Timestamp:    time.Now(),
		CPUUsage:     0.5,
		MemoryUsage:  0.3,
		DiskUsage:    0.2,
		RequestCount: 0,
		ErrorCount:   0,
		AvgLatency:   0,
	}, nil
}

// ExportMetrics exports metrics in Prometheus format
func (mc *EtcdMetricsCollector) ExportMetrics() ([]byte, error) {
	// TODO: Implement actual Prometheus metrics export
	return []byte("# VJVector metrics placeholder"), nil
}

// EtcdConsensusProtocol implements ConsensusProtocol using etcd
type EtcdConsensusProtocol struct {
	client *clientv3.Client
	logger *slog.Logger
}

// Start starts the consensus protocol
func (cp *EtcdConsensusProtocol) Start(ctx context.Context) error {
	cp.logger.Info("Starting consensus protocol")
	// TODO: Implement actual consensus protocol start
	return nil
}

// Stop stops the consensus protocol
func (cp *EtcdConsensusProtocol) Stop(ctx context.Context) error {
	cp.logger.Info("Stopping consensus protocol")
	// TODO: Implement actual consensus protocol stop
	return nil
}

// Propose proposes a value for consensus
func (cp *EtcdConsensusProtocol) Propose(ctx context.Context, value interface{}) error {
	// TODO: Implement actual consensus proposal
	cp.logger.Info("Proposing value for consensus", "value", value)
	return nil
}

// GetValue returns the current consensus value
func (cp *EtcdConsensusProtocol) GetValue(ctx context.Context) (interface{}, error) {
	// TODO: Implement actual consensus value retrieval
	return "consensus_value_placeholder", nil
}

// GetLeader returns the current leader
func (cp *EtcdConsensusProtocol) GetLeader() string {
	// TODO: Implement actual leader retrieval
	return "leader_placeholder"
}

// IsLeader returns true if this node is the leader
func (cp *EtcdConsensusProtocol) IsLeader() bool {
	// TODO: Implement actual leader check
	return false
}

// RoundRobinLoadBalancer implements round-robin load balancing
type RoundRobinLoadBalancer struct {
	mu       sync.RWMutex
	nodes    []*NodeInfo
	current  int
	requests int64
}

// NewRoundRobinLoadBalancer creates a new round-robin load balancer
func NewRoundRobinLoadBalancer() LoadBalancer {
	return &RoundRobinLoadBalancer{
		nodes:   make([]*NodeInfo, 0),
		current: 0,
	}
}

// GetNode returns the next node in round-robin fashion
func (rr *RoundRobinLoadBalancer) GetNode(request *LoadBalancerRequest) (*NodeInfo, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(rr.nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	node := rr.nodes[rr.current]
	rr.current = (rr.current + 1) % len(rr.nodes)
	rr.requests++

	return node, nil
}

// UpdateNode updates node information
func (rr *RoundRobinLoadBalancer) UpdateNode(nodeInfo *NodeInfo) error {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	// Find and update existing node
	for i, node := range rr.nodes {
		if node.ID == nodeInfo.ID {
			rr.nodes[i] = nodeInfo
			return nil
		}
	}

	// Add new node
	rr.nodes = append(rr.nodes, nodeInfo)
	return nil
}

// RemoveNode removes a node from the load balancer
func (rr *RoundRobinLoadBalancer) RemoveNode(nodeID string) error {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	for i, node := range rr.nodes {
		if node.ID == nodeID {
			rr.nodes = append(rr.nodes[:i], rr.nodes[i+1:]...)
			if rr.current >= len(rr.nodes) {
				rr.current = 0
			}
			return nil
		}
	}

	return fmt.Errorf("node not found: %s", nodeID)
}

// GetStats returns load balancing statistics
func (rr *RoundRobinLoadBalancer) GetStats() *LoadBalancerStats {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	nodeLoads := make(map[string]float64)
	for _, node := range rr.nodes {
		nodeLoads[node.ID] = 0.0 // TODO: Implement actual load calculation
	}

	return &LoadBalancerStats{
		TotalRequests:  rr.requests,
		ActiveRequests: 0, // TODO: Implement active request tracking
		NodeLoads:      nodeLoads,
		ResponseTimes:  make(map[string]time.Duration),
		LastUpdated:    time.Now(),
	}
}
