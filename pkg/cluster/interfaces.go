package cluster

import (
	"context"
	"time"
)

// NodeRole represents the role of a cluster node
type NodeRole string

const (
	NodeRoleMaster    NodeRole = "master"
	NodeRoleSlave     NodeRole = "slave"
	NodeRoleCandidate NodeRole = "candidate"
)

// NodeState represents the current state of a cluster node
type NodeState string

const (
	NodeStateStarting   NodeState = "starting"
	NodeStateRunning    NodeState = "running"
	NodeStateStopping   NodeState = "stopping"
	NodeStateStopped    NodeState = "stopped"
	NodeStateFailed     NodeState = "failed"
	NodeStateRecovering NodeState = "recovering"
)

// NodeInfo represents information about a cluster node
type NodeInfo struct {
	ID        string                 `json:"id"`
	Address   string                 `json:"address"`
	Port      int                    `json:"port"`
	Role      NodeRole               `json:"role"`
	State     NodeState              `json:"state"`
	Version   string                 `json:"version"`
	StartTime time.Time              `json:"start_time"`
	LastSeen  time.Time              `json:"last_seen"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Peer represents a peer node in the cluster
type Peer struct {
	Info       *NodeInfo
	Connection Connection
	LastPing   time.Time
	Latency    time.Duration
}

// Connection represents a connection to a peer node
type Connection interface {
	// Send sends a message to the peer
	Send(ctx context.Context, message *Message) error

	// Receive receives a message from the peer
	Receive(ctx context.Context) (*Message, error)

	// Close closes the connection
	Close() error

	// IsConnected returns true if the connection is active
	IsConnected() bool
}

// Message represents a message sent between cluster nodes
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Sequence  uint64                 `json:"sequence"`
}

// MessageType represents the type of cluster message
type MessageType string

const (
	MessageTypePing      MessageType = "ping"
	MessageTypePong      MessageType = "pong"
	MessageTypeHeartbeat MessageType = "heartbeat"
	MessageTypeElection  MessageType = "election"
	MessageTypeVote      MessageType = "vote"
	MessageTypeSync      MessageType = "sync"
	MessageTypeReplicate MessageType = "replicate"
	MessageTypeShard     MessageType = "shard"
	MessageTypeHealth    MessageType = "health"
	MessageTypeConfig    MessageType = "config"
)

// Cluster represents a cluster of VJVector nodes
type Cluster interface {
	// Start starts the cluster
	Start(ctx context.Context) error

	// Stop stops the cluster
	Stop(ctx context.Context) error

	// GetNodeInfo returns information about the current node
	GetNodeInfo() *NodeInfo

	// GetPeers returns information about peer nodes
	GetPeers() []*Peer

	// GetMaster returns the current master node
	GetMaster() *NodeInfo

	// IsMaster returns true if this node is the master
	IsMaster() bool

	// Join joins an existing cluster
	Join(ctx context.Context, seedNode string) error

	// Leave leaves the cluster
	Leave(ctx context.Context) error

	// Health returns the health status of the cluster
	Health(ctx context.Context) (*ClusterHealth, error)
}

// ClusterHealth represents the health status of the cluster
type ClusterHealth struct {
	Status      string                 `json:"status"`
	NodeCount   int                    `json:"node_count"`
	MasterID    string                 `json:"master_id"`
	Replicas    int                    `json:"replicas"`
	Shards      int                    `json:"shards"`
	LastUpdated time.Time              `json:"last_updated"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ConsensusProtocol represents the consensus protocol for cluster coordination
type ConsensusProtocol interface {
	// Start starts the consensus protocol
	Start(ctx context.Context) error

	// Stop stops the consensus protocol
	Stop(ctx context.Context) error

	// Propose proposes a value for consensus
	Propose(ctx context.Context, value interface{}) error

	// GetValue returns the current consensus value
	GetValue(ctx context.Context) (interface{}, error)

	// GetLeader returns the current leader
	GetLeader() string

	// IsLeader returns true if this node is the leader
	IsLeader() bool
}

// ShardingStrategy represents the strategy for data sharding
type ShardingStrategy interface {
	// GetShard returns the shard for a given key
	GetShard(key string) int

	// GetShardCount returns the total number of shards
	GetShardCount() int

	// AddShard adds a new shard
	AddShard() error

	// RemoveShard removes a shard
	RemoveShard(shardID int) error

	// Rebalance rebalances data across shards
	Rebalance() error
}

// LoadBalancer represents the load balancing strategy
type LoadBalancer interface {
	// GetNode returns the best node for a given request
	GetNode(request *LoadBalancerRequest) (*NodeInfo, error)

	// UpdateNode updates node information
	UpdateNode(nodeInfo *NodeInfo) error

	// RemoveNode removes a node from the load balancer
	RemoveNode(nodeID string) error

	// GetStats returns load balancing statistics
	GetStats() *LoadBalancerStats
}

// LoadBalancerRequest represents a request for load balancing
type LoadBalancerRequest struct {
	Operation string                 `json:"operation"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Priority  int                    `json:"priority"`
	UserID    string                 `json:"user_id,omitempty"`
}

// LoadBalancerStats represents load balancing statistics
type LoadBalancerStats struct {
	TotalRequests  int64                    `json:"total_requests"`
	ActiveRequests int64                    `json:"active_requests"`
	NodeLoads      map[string]float64       `json:"node_loads"`
	ResponseTimes  map[string]time.Duration `json:"response_times"`
	LastUpdated    time.Time                `json:"last_updated"`
}

// PeerManager manages peer connections
type PeerManager interface {
	// AddPeer adds a new peer
	AddPeer(peer *Peer) error

	// RemovePeer removes a peer
	RemovePeer(peerID string) error

	// GetPeer returns a peer by ID
	GetPeer(peerID string) (*Peer, error)

	// GetPeers returns all peers
	GetPeers() []*Peer

	// Broadcast broadcasts a message to all peers
	Broadcast(ctx context.Context, message *Message) error

	// SendToPeer sends a message to a specific peer
	SendToPeer(ctx context.Context, peerID string, message *Message) error
}

// ReplicationManager manages data replication
type ReplicationManager interface {
	// StartReplication starts replication for a shard
	StartReplication(ctx context.Context, shardID int) error

	// StopReplication stops replication for a shard
	StopReplication(ctx context.Context, shardID int) error

	// GetReplicationStatus returns the replication status
	GetReplicationStatus(shardID int) (*ReplicationStatus, error)

	// SyncData synchronizes data with peers
	SyncData(ctx context.Context, shardID int) error

	// GetReplicaNodes returns nodes that have replicas of a shard
	GetReplicaNodes(shardID int) []*NodeInfo
}

// ReplicationStatus represents the status of data replication
type ReplicationStatus struct {
	ShardID      int           `json:"shard_id"`
	Status       string        `json:"status"`
	LastSync     time.Time     `json:"last_sync"`
	SyncLag      time.Duration `json:"sync_lag"`
	ReplicaCount int           `json:"replica_count"`
	ReplicaNodes []string      `json:"replica_nodes"`
	LastError    string        `json:"last_error,omitempty"`
}

// HealthChecker checks the health of cluster components
type HealthChecker interface {
	// CheckHealth checks the health of the cluster
	CheckHealth(ctx context.Context) (*ClusterHealth, error)

	// CheckNodeHealth checks the health of a specific node
	CheckNodeHealth(ctx context.Context, nodeID string) (*NodeHealth, error)

	// RegisterHealthCheck registers a custom health check
	RegisterHealthCheck(name string, check HealthCheckFunc) error
}

// NodeHealth represents the health of a specific node
type NodeHealth struct {
	NodeID       string                 `json:"node_id"`
	Status       string                 `json:"status"`
	LastCheck    time.Time              `json:"last_check"`
	ResponseTime time.Duration          `json:"response_time"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// HealthCheckFunc represents a custom health check function
type HealthCheckFunc func(ctx context.Context) (bool, error)

// MetricsCollector collects cluster metrics
type MetricsCollector interface {
	// CollectMetrics collects metrics from the cluster
	CollectMetrics(ctx context.Context) (*ClusterMetrics, error)

	// GetNodeMetrics returns metrics for a specific node
	GetNodeMetrics(ctx context.Context, nodeID string) (*NodeMetrics, error)

	// ExportMetrics exports metrics in Prometheus format
	ExportMetrics() ([]byte, error)
}

// ClusterMetrics represents cluster-wide metrics
type ClusterMetrics struct {
	Timestamp      time.Time              `json:"timestamp"`
	NodeCount      int                    `json:"node_count"`
	ActiveNodes    int                    `json:"active_nodes"`
	TotalRequests  int64                  `json:"total_requests"`
	ActiveRequests int64                  `json:"active_requests"`
	AvgLatency     time.Duration          `json:"avg_latency"`
	ErrorRate      float64                `json:"error_rate"`
	ShardCount     int                    `json:"shard_count"`
	ReplicaCount   int                    `json:"replica_count"`
	Details        map[string]interface{} `json:"details,omitempty"`
}

// NodeMetrics represents metrics for a specific node
type NodeMetrics struct {
	NodeID       string                 `json:"node_id"`
	Timestamp    time.Time              `json:"timestamp"`
	CPUUsage     float64                `json:"cpu_usage"`
	MemoryUsage  float64                `json:"memory_usage"`
	DiskUsage    float64                `json:"disk_usage"`
	NetworkIO    *NetworkMetrics        `json:"network_io"`
	RequestCount int64                  `json:"request_count"`
	ErrorCount   int64                  `json:"error_count"`
	AvgLatency   time.Duration          `json:"avg_latency"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// NetworkMetrics represents network I/O metrics
type NetworkMetrics struct {
	BytesReceived   int64 `json:"bytes_received"`
	BytesSent       int64 `json:"bytes_sent"`
	PacketsReceived int64 `json:"packets_received"`
	PacketsSent     int64 `json:"packets_sent"`
	Errors          int64 `json:"errors"`
}
