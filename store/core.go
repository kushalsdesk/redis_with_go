package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"
)

type ValueType int

const (
	STRING ValueType = iota
	LIST
	STREAM
)

type StreamEntry struct {
	ID     string
	Fields map[string]string
}

type Stream struct {
	Entries []StreamEntry
	LastID  string
}

type StreamID struct {
	Timestamp int64
	Sequence  int64
}

type RedisValue struct {
	Type   ValueType
	String string
	List   []string
	Stream *Stream
	Expiry *time.Time
}

type ReplicationState struct {
	Role             string
	MasterHost       string
	MasterPort       string
	MasterReplID     string
	MasterReplOffset int64
	ConnectedSlaves  int
	Replicas         []string
	ReplicaConns     map[string]*ReplicationConnection
	SlaveOffset      int64
}

type ReplicationConnection struct {
	Address    string
	Connection net.Conn
	Connected  bool
	Offset     int64
	LastACK    time.Time
	Lag        int64
	ReplID     string
}

var (
	data             = make(map[string]*RedisValue)
	dataMutex        sync.RWMutex
	replicationState = &ReplicationState{
		Role:             "master",
		MasterReplID:     generateReplID(),
		MasterReplOffset: 0,
		ConnectedSlaves:  0,
		Replicas:         make([]string, 0),
		ReplicaConns:     make(map[string]*ReplicationConnection),
	}
	replicationMutex sync.RWMutex
)

func getReplicaID(conn net.Conn) string {
	return conn.RemoteAddr().String()

}

func generateReplID() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func AddReplicaWithConnection(conn net.Conn) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()

	address := conn.RemoteAddr().String()
	replicationState.Replicas = append(replicationState.Replicas, address)
	replicationState.ReplicaConns[address] = &ReplicationConnection{
		Address:    address,
		Connection: conn,
		Connected:  true,
	}
	replicationState.ConnectedSlaves = len(replicationState.ReplicaConns)
	fmt.Printf("🔗 Replica connected: %s\n", address)
}

func RemoveReplicaByConnection(conn net.Conn) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()

	address := conn.RemoteAddr().String()

	// Remove from replica list
	for i, replica := range replicationState.Replicas {
		if replica == address {
			replicationState.Replicas = append(replicationState.Replicas[:i], replicationState.Replicas[i+1:]...)
			break
		}
	}

	// Remove from connection map
	if replica, exists := replicationState.ReplicaConns[address]; exists {
		replica.Connected = false
		delete(replicationState.ReplicaConns, address)
	}

	replicationState.ConnectedSlaves = len(replicationState.ReplicaConns)
	fmt.Printf("❌ Replica disconnected: %s\n", address)
}

func GetReplicaConnections() []*ReplicationConnection {
	replicationMutex.RLock()
	defer replicationMutex.RUnlock()

	connections := make([]*ReplicationConnection, 0, len(replicationState.ReplicaConns))
	for _, replica := range replicationState.ReplicaConns {
		if replica.Connected {
			connections = append(connections, replica)
		}
	}
	return connections
}

func GetReplicationState() *ReplicationState {
	replicationMutex.RLock()
	defer replicationMutex.RUnlock()

	replicaConnsCopy := make(map[string]*ReplicationConnection)
	for k, v := range replicationState.ReplicaConns {
		replicaConnsCopy[k] = v
	}

	return &ReplicationState{
		Role:             replicationState.Role,
		MasterHost:       replicationState.MasterHost,
		MasterPort:       replicationState.MasterPort,
		MasterReplID:     replicationState.MasterReplID,
		MasterReplOffset: replicationState.MasterReplOffset,
		ConnectedSlaves:  replicationState.ConnectedSlaves,
		Replicas:         append([]string{}, replicationState.Replicas...),
		ReplicaConns:     replicaConnsCopy,
		SlaveOffset:      replicationState.SlaveOffset,
	}
}

func SetReplicationRole(role, masterHost, masterPort string) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()

	replicationState.Role = role
	replicationState.MasterHost = masterHost
	replicationState.MasterPort = masterPort

	if role == "slave" {
		replicationState.ConnectedSlaves = 0
		replicationState.Replicas = make([]string, 0)
		replicationState.SlaveOffset = 0
		InitACKChannel()
	} else {
		CloseACKChannel()
	}
}

func IncrementReplOffset(increment int64) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()
	replicationState.MasterReplOffset += increment
}

func GetReplOffset() int64 {
	replicationMutex.RLock()
	defer replicationMutex.RUnlock()
	return replicationState.MasterReplOffset
}

func AddReplica(replica string) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()
	replicationState.Replicas = append(replicationState.Replicas, replica)
	replicationState.ConnectedSlaves = len(replicationState.Replicas)
}

func RemoveReplica(replica string) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()

	for i, rep := range replicationState.Replicas {
		if rep == replica {
			replicationState.Replicas = append(replicationState.Replicas[:i], replicationState.Replicas[i+1:]...)
			break
		}
	}
	replicationState.ConnectedSlaves = len(replicationState.Replicas)
}

func GetKeyType(key string) string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	value, exists := data[key]
	if !exists {
		return "none"
	}

	if isExpired(value, key) {
		return "none"
	}

	switch value.Type {
	case STRING:
		return "string"
	case LIST:
		return "list"
	case STREAM:
		return "stream"
	default:
		return "none"
	}
}

func isExpired(value *RedisValue, key string) bool {
	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		delete(data, key)
		return true
	}
	return false
}
