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

// EtcdCluster implements the Cluster interface using etcd for coordination
type EtcdCluster struct {
	mu sync.RWMutex

	// Node information
	nodeInfo *NodeInfo

	// etcd client
	etcdClient *clientv3.Client

	// Cluster state
	peers  map[string]*Peer
	master *NodeInfo
	role   NodeRole
	state  NodeState

	// Configuration
	config *ClusterConfig

	// Managers
	peerManager        PeerManager
	replicationManager ReplicationManager
	healthChecker      HealthChecker
	metricsCollector   MetricsCollector

	// Consensus and coordination
	consensus    ConsensusProtocol
	sharding     ShardingStrategy
	loadBalancer LoadBalancer

	// Internal state
	ctx             context.Context
	cancel          context.CancelFunc
	heartbeatTicker *time.Ticker
	electionTicker  *time.Ticker

	// Logging
	logger *slog.Logger
}

// ClusterConfig holds cluster configuration
type ClusterConfig struct {
	// Node configuration
	NodeID  string   `json:"node_id"`
	Address string   `json:"address"`
	Port    int      `json:"port"`
	Role    NodeRole `json:"role"`

	// etcd configuration
	EtcdEndpoints []string      `json:"etcd_endpoints"`
	EtcdTimeout   time.Duration `json:"etcd_timeout"`

	// Clustering configuration
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	ElectionTimeout   time.Duration `json:"election_timeout"`
	MaxPeers          int           `json:"max_peers"`

	// Sharding configuration
	ShardCount   int `json:"shard_count"`
	ReplicaCount int `json:"replica_count"`

	// Load balancing configuration
	LoadBalancingStrategy string `json:"load_balancing_strategy"`
}

// NewEtcdCluster creates a new etcd-based cluster
func NewEtcdCluster(config *ClusterConfig, logger *slog.Logger) (*EtcdCluster, error) {
	if config.NodeID == "" {
		// Generate random node ID if not provided
		id, err := generateNodeID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate node ID: %w", err)
		}
		config.NodeID = id
	}

	// Set defaults
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 5 * time.Second
	}
	if config.ElectionTimeout == 0 {
		config.ElectionTimeout = 10 * time.Second
	}
	if config.MaxPeers == 0 {
		config.MaxPeers = 10
	}
	if config.ShardCount == 0 {
		config.ShardCount = 8
	}
	if config.ReplicaCount == 0 {
		config.ReplicaCount = 3
	}
	if config.LoadBalancingStrategy == "" {
		config.LoadBalancingStrategy = "round_robin"
	}

	cluster := &EtcdCluster{
		config: config,
		peers:  make(map[string]*Peer),
		state:  NodeStateStarting,
		logger: logger,
	}

	// Initialize node info
	cluster.nodeInfo = &NodeInfo{
		ID:        config.NodeID,
		Address:   config.Address,
		Port:      config.Port,
		Role:      config.Role,
		State:     NodeStateStarting,
		Version:   "1.0.0", // TODO: Get from build info
		StartTime: time.Now(),
		LastSeen:  time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	return cluster, nil
}

// Start starts the cluster
func (c *EtcdCluster) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != NodeStateStarting {
		return fmt.Errorf("cluster is not in starting state: %s", c.state)
	}

	c.logger.Info("Starting etcd cluster",
		"node_id", c.nodeInfo.ID,
		"address", c.nodeInfo.Address,
		"port", c.nodeInfo.Port,
		"role", c.nodeInfo.Role)

	// Create etcd client
	etcdConfig := clientv3.Config{
		Endpoints:   c.config.EtcdEndpoints,
		DialTimeout: c.config.EtcdTimeout,
	}

	etcdClient, err := clientv3.New(etcdConfig)
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %w", err)
	}
	c.etcdClient = etcdClient

	// Test etcd connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = c.etcdClient.Get(ctx, "test")
	if err != nil && err.Error() != "etcdserver: key not found" {
		return fmt.Errorf("failed to connect to etcd: %w", err)
	}

	// Initialize managers
	if err := c.initializeManagers(); err != nil {
		return fmt.Errorf("failed to initialize managers: %w", err)
	}

	// Create context for background operations
	c.ctx, c.cancel = context.WithCancel(context.Background())

	// Start background operations
	go c.startHeartbeat()
	go c.startElection()
	go c.watchClusterChanges()

	// Update state
	c.state = NodeStateRunning
	c.nodeInfo.State = NodeStateRunning

	c.logger.Info("Etcd cluster started successfully",
		"node_id", c.nodeInfo.ID,
		"state", c.state)

	return nil
}

// Stop stops the cluster
func (c *EtcdCluster) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == NodeStateStopped {
		return nil
	}

	c.logger.Info("Stopping etcd cluster", "node_id", c.nodeInfo.ID)

	// Update state
	c.state = NodeStateStopping
	c.nodeInfo.State = NodeStateStopping

	// Cancel background operations
	if c.cancel != nil {
		c.cancel()
	}

	// Stop tickers
	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
	}
	if c.electionTicker != nil {
		c.electionTicker.Stop()
	}

	// Close etcd client
	if c.etcdClient != nil {
		if err := c.etcdClient.Close(); err != nil {
			c.logger.Error("Failed to close etcd client", "error", err)
		}
	}

	// Update state
	c.state = NodeStateStopped
	c.nodeInfo.State = NodeStateStopped

	c.logger.Info("Etcd cluster stopped", "node_id", c.nodeInfo.ID)

	return nil
}

// GetNodeInfo returns information about the current node
func (c *EtcdCluster) GetNodeInfo() *NodeInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid race conditions
	info := *c.nodeInfo
	info.LastSeen = time.Now()
	return &info
}

// GetPeers returns information about peer nodes
func (c *EtcdCluster) GetPeers() []*Peer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	peers := make([]*Peer, 0, len(c.peers))
	for _, peer := range c.peers {
		peers = append(peers, peer)
	}
	return peers
}

// GetMaster returns the current master node
func (c *EtcdCluster) GetMaster() *NodeInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.master == nil {
		return nil
	}

	// Return a copy to avoid race conditions
	master := *c.master
	return &master
}

// IsMaster returns true if this node is the master
func (c *EtcdCluster) IsMaster() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.role == NodeRoleMaster
}

// Join joins an existing cluster
func (c *EtcdCluster) Join(ctx context.Context, seedNode string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != NodeStateRunning {
		return fmt.Errorf("cluster is not running: %s", c.state)
	}

	c.logger.Info("Joining cluster", "seed_node", seedNode, "node_id", c.nodeInfo.ID)

	// Register this node in etcd
	nodeKey := fmt.Sprintf("/vjvector/nodes/%s", c.nodeInfo.ID)
	nodeData, err := json.Marshal(c.nodeInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal node info: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = c.etcdClient.Put(ctx, nodeKey, string(nodeData))
	if err != nil {
		return fmt.Errorf("failed to register node in etcd: %w", err)
	}

	// Discover existing nodes
	if err := c.discoverNodes(ctx); err != nil {
		return fmt.Errorf("failed to discover nodes: %w", err)
	}

	c.logger.Info("Successfully joined cluster", "node_id", c.nodeInfo.ID)

	return nil
}

// Leave leaves the cluster
func (c *EtcdCluster) Leave(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != NodeStateRunning {
		return nil
	}

	c.logger.Info("Leaving cluster", "node_id", c.nodeInfo.ID)

	// Remove this node from etcd
	nodeKey := fmt.Sprintf("/vjvector/nodes/%s", c.nodeInfo.ID)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.etcdClient.Delete(ctx, nodeKey)
	if err != nil {
		c.logger.Error("Failed to remove node from etcd", "error", err)
	}

	// Close peer connections
	for _, peer := range c.peers {
		if peer.Connection != nil {
			peer.Connection.Close()
		}
	}
	c.peers = make(map[string]*Peer)

	c.logger.Info("Successfully left cluster", "node_id", c.nodeInfo.ID)

	return nil
}

// Health returns the health status of the cluster
func (c *EtcdCluster) Health(ctx context.Context) (*ClusterHealth, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	health := &ClusterHealth{
		Status:      "healthy",
		NodeCount:   len(c.peers) + 1, // +1 for self
		MasterID:    "",
		Replicas:    c.config.ReplicaCount,
		Shards:      c.config.ShardCount,
		LastUpdated: time.Now(),
		Details:     make(map[string]interface{}),
	}

	if c.master != nil {
		health.MasterID = c.master.ID
	}

	// Check if we have enough peers
	if len(c.peers) < c.config.ReplicaCount-1 {
		health.Status = "degraded"
		health.Details["warning"] = "insufficient replicas"
	}

	// Check node health
	unhealthyNodes := 0
	for _, peer := range c.peers {
		if time.Since(peer.LastPing) > c.config.HeartbeatInterval*3 {
			unhealthyNodes++
		}
	}

	if unhealthyNodes > 0 {
		health.Status = "degraded"
		health.Details["unhealthy_nodes"] = unhealthyNodes
	}

	return health, nil
}

// initializeManagers initializes cluster managers
func (c *EtcdCluster) initializeManagers() error {
	// Initialize peer manager
	c.peerManager = NewEtcdPeerManager(c.etcdClient, c.config, c.logger)

	// Initialize replication manager
	c.replicationManager = NewEtcdReplicationManager(c.etcdClient, c.config, c.logger)

	// Initialize health checker
	c.healthChecker = NewEtcdHealthChecker(c.etcdClient, c.config, c.logger)

	// Initialize metrics collector
	c.metricsCollector = NewEtcdMetricsCollector(c.etcdClient, c.config, c.logger)

	// Initialize consensus protocol
	c.consensus = NewEtcdConsensusProtocol(c.etcdClient, c.config, c.logger)

	// Initialize sharding strategy
	c.sharding = NewHashSharding(c.config.ShardCount)

	// Initialize load balancer
	var err error
	c.loadBalancer, err = NewLoadBalancer(c.config.LoadBalancingStrategy, c.config, c.logger)
	if err != nil {
		return fmt.Errorf("failed to create load balancer: %w", err)
	}

	return nil
}

// startHeartbeat starts the heartbeat mechanism
func (c *EtcdCluster) startHeartbeat() {
	c.heartbeatTicker = time.NewTicker(c.config.HeartbeatInterval)
	defer c.heartbeatTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.heartbeatTicker.C:
			c.sendHeartbeat()
		}
	}
}

// startElection starts the leader election mechanism
func (c *EtcdCluster) startElection() {
	c.electionTicker = time.NewTicker(c.config.ElectionTimeout)
	defer c.electionTicker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.electionTicker.C:
			c.runElection()
		}
	}
}

// sendHeartbeat sends heartbeat to etcd
func (c *EtcdCluster) sendHeartbeat() {
	c.mu.RLock()
	if c.state != NodeStateRunning {
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()

	// Update last seen time
	c.mu.Lock()
	c.nodeInfo.LastSeen = time.Now()
	c.mu.Unlock()

	// Send heartbeat to etcd
	heartbeatKey := fmt.Sprintf("/vjvector/heartbeats/%s", c.nodeInfo.ID)
	heartbeatData := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"state":     c.state,
		"role":      c.role,
	}

	heartbeatBytes, err := json.Marshal(heartbeatData)
	if err != nil {
		c.logger.Error("Failed to marshal heartbeat", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = c.etcdClient.Put(ctx, heartbeatKey, string(heartbeatBytes))
	if err != nil {
		c.logger.Error("Failed to send heartbeat", "error", err)
	}
}

// runElection runs the leader election process
func (c *EtcdCluster) runElection() {
	c.mu.RLock()
	if c.state != NodeStateRunning || c.role == NodeRoleMaster {
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()

	// Try to become master
	if err := c.tryBecomeMaster(); err != nil {
		c.logger.Debug("Failed to become master", "error", err)
	}
}

// tryBecomeMaster attempts to become the master node
func (c *EtcdCluster) tryBecomeMaster() error {
	masterKey := "/vjvector/master"
	masterData := map[string]interface{}{
		"node_id":   c.nodeInfo.ID,
		"timestamp": time.Now().Unix(),
		"term":      time.Now().UnixNano(),
	}

	masterBytes, err := json.Marshal(masterData)
	if err != nil {
		return fmt.Errorf("failed to marshal master data: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try to create the master key (only succeeds if it doesn't exist)
	resp, err := c.etcdClient.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(masterKey), "=", 0)).
		Then(clientv3.OpPut(masterKey, string(masterBytes))).
		Commit()

	if err != nil {
		return fmt.Errorf("failed to run master election transaction: %w", err)
	}

	if resp.Succeeded {
		// We became the master
		c.mu.Lock()
		c.role = NodeRoleMaster
		c.nodeInfo.Role = NodeRoleMaster
		c.master = c.nodeInfo
		c.mu.Unlock()

		c.logger.Info("Became master node", "node_id", c.nodeInfo.ID)
	}

	return nil
}

// watchClusterChanges watches for cluster changes in etcd
func (c *EtcdCluster) watchClusterChanges() {
	watchChan := c.etcdClient.Watch(context.Background(), "/vjvector/", clientv3.WithPrefix())

	for {
		select {
		case <-c.ctx.Done():
			return
		case watchResp := <-watchChan:
			if watchResp.Err() != nil {
				c.logger.Error("Watch error", "error", watchResp.Err())
				continue
			}

			for _, event := range watchResp.Events {
				c.handleClusterEvent(event)
			}
		}
	}
}

// handleClusterEvent handles cluster events from etcd
func (c *EtcdCluster) handleClusterEvent(event *clientv3.Event) {
	switch event.Type {
	case clientv3.EventTypePut:
		c.handlePutEvent(event)
	case clientv3.EventTypeDelete:
		c.handleDeleteEvent(event)
	}
}

// handlePutEvent handles PUT events from etcd
func (c *EtcdCluster) handlePutEvent(event *clientv3.Event) {
	key := string(event.Kv.Key)

	if key == "/vjvector/master" {
		c.handleMasterChange(event.Kv.Value)
	} else if len(key) > 13 && key[:13] == "/vjvector/nodes/" {
		c.handleNodeJoin(key, event.Kv.Value)
	}
}

// handleDeleteEvent handles DELETE events from etcd
func (c *EtcdCluster) handleDeleteEvent(event *clientv3.Event) {
	key := string(event.Kv.Key)

	if len(key) > 13 && key[:13] == "/vjvector/nodes/" {
		nodeID := key[13:]
		c.handleNodeLeave(nodeID)
	}
}

// handleMasterChange handles master node changes
func (c *EtcdCluster) handleMasterChange(value []byte) {
	var masterData map[string]interface{}
	if err := json.Unmarshal(value, &masterData); err != nil {
		c.logger.Error("Failed to unmarshal master data", "error", err)
		return
	}

	masterID, ok := masterData["node_id"].(string)
	if !ok {
		c.logger.Error("Invalid master data format")
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if masterID == c.nodeInfo.ID {
		// We are the master
		c.role = NodeRoleMaster
		c.nodeInfo.Role = NodeRoleMaster
		c.master = c.nodeInfo
	} else {
		// Another node is the master
		c.role = NodeRoleSlave
		c.nodeInfo.Role = NodeRoleSlave
		c.master = &NodeInfo{ID: masterID}
	}

	c.logger.Info("Master changed", "new_master", masterID, "our_role", c.role)
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
			peer.Connection.Close()
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
