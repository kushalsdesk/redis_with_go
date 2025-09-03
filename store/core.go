package store

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type ValueType int

const (
	STRING ValueType = iota
	LIST
	STREAM
)

// stream entry structure
type StreamEntry struct {
	ID     string
	Fields map[string]string
}

// stream data structure
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
}

var (
	data      = make(map[string]*RedisValue)
	dataMutex sync.RWMutex

	// Replication state
	replicationState = &ReplicationState{
		Role:             "master",
		MasterReplID:     generateReplID(),
		MasterReplOffset: 0,
		ConnectedSlaves:  0,
		Replicas:         make([]string, 0),
	}
	replicationMutex sync.RWMutex
)

func generateReplID() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func GetReplicationState() *ReplicationState {
	replicationMutex.RLock()
	defer replicationMutex.RUnlock()

	// Returning a copy to avoid conditions
	return &ReplicationState{
		Role:             replicationState.Role,
		MasterHost:       replicationState.MasterHost,
		MasterPort:       replicationState.MasterPort,
		MasterReplID:     replicationState.MasterReplID,
		MasterReplOffset: replicationState.MasterReplOffset,
		ConnectedSlaves:  replicationState.ConnectedSlaves,
		Replicas:         append([]string{}, replicationState.Replicas...),
	}
}

func SetReplicationRole(role, masterHost, masterPort string) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()

	replicationState.Role = role
	replicationState.MasterHost = masterHost
	replicationState.MasterPort = masterPort

	if role == "slave" {
		// Reset master- specific state when becoming a replica
		replicationState.ConnectedSlaves = 0
		replicationState.Replicas = make([]string, 0)
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

	for index, rep := range replicationState.Replicas {
		if rep == replica {
			replicationState.Replicas = append(replicationState.Replicas[:index], replicationState.Replicas[index+1:]...)
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

	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		delete(data, key)
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
