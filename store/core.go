package store

import (
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

var (
	data      = make(map[string]*RedisValue)
	dataMutex sync.RWMutex
)

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
