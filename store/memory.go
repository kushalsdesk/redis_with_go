package store

import (
	"sync"
	"time"
)

type ValueType int

const (
	STRING ValueType = iota
	LIST
)

type RedisValue struct {
	Type   ValueType
	String string
	List   []string
	Expiry *time.Time
}

var (
	data      = make(map[string]*RedisValue)
	dataMutex sync.RWMutex
)

// BlockingClient represents a client for list elements
type BlockingClient struct {
	Keys     []string
	Left     bool
	Response chan BlockingResult
	Timeout  time.Duration
}

type BlockingResult struct {
	Key     string
	Value   string
	Success bool
}

var (
	blockingClients = make(map[string][]*BlockingClient)
	blockingMutex   sync.Mutex
)

func Set(key, val string, ttl time.Duration) {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	value := &RedisValue{
		Type:   STRING,
		String: val,
	}

	if ttl > 0 {
		expiry := time.Now().Add(ttl)
		value.Expiry = &expiry
	}

	data[key] = value

}

func Get(key string) (string, bool) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	value, exists := data[key]
	if !exists {
		return "", false
	}

	//check expiry
	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		delete(data, key)
		return "", false
	}

	//check type
	if value.Type != STRING {
		return "", false
	}

	return value.String, true
}

func GetListLength(key string) int {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	value, exists := data[key]
	if !exists {
		return 0
	}

	if value.Type != LIST {
		return -1
	}

	//check expiry
	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		return 0
	}

	return len(value.List)

}

// Get elements at specific index (both positive and negative indexes)

func ListIndex(key string, index int) (string, bool) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	value, exists := data[key]

	if !exists {
		return "", false
	}

	if value.Type != LIST {
		return "", false
	}

	//check expiry
	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		return "", false
	}

	listlen := len(value.List)
	if listlen == 0 {
		return "", false
	}

	// handle negative indexing
	if index < 0 {
		index = listlen + index
	}

	//check bounds
	if index < 0 || index >= listlen {
		return "", false
	}

	return value.List[index], true

}

// Get range of elements from list
//

func ListRange(key string, start, stop int) ([]string, bool) {

	dataMutex.RLock()
	defer dataMutex.RUnlock()

	value, exists := data[key]
	if !exists {
		return []string{}, true
	}

	if value.Type != LIST {
		return nil, false
	}

	//Check expiry
	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		return []string{}, true
	}

	listlen := len(value.List)
	if listlen == 0 {
		return []string{}, true
	}

	// handle negative indexing
	if start < 0 {
		start = listlen + start
	}
	if stop < 0 {
		stop = listlen + stop
	}

	// clamp to bounds
	if start < 0 {
		start = 0
	}
	if stop >= listlen {
		stop = listlen - 1
	}

	// if start > stop , return empty
	if start > stop {
		return []string{}, true
	}

	return value.List[start : stop+1], true
}

// Keeping the single pop function for backward compatibility
func ListPop(key string, left bool) (string, bool) {
	elements, ok := ListPopMultiple(key, 1, left)
	if !ok || len(elements) == 0 {
		return "", false
	}
	return elements[0], true
}

// Remove and return multiple elements from list

func ListPopMultiple(key string, count int, left bool) ([]string, bool) {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	value, exists := data[key]
	if !exists {
		return nil, false
	}

	if value.Type != LIST {
		return nil, false
	}

	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		delete(data, key)
		return nil, false
	}

	listlen := len(value.List)
	if listlen == 0 {
		return []string{}, true
	}

	if count > listlen {
		count = listlen
	}

	var result []string
	if left {
		result = make([]string, count)
		copy(result, value.List[:count])
		value.List = value.List[count:]
	} else {
		result = make([]string, count)
		startIndex := listlen - count
		copy(result, value.List[startIndex:])
		value.List = value.List[:startIndex]

		for i := 0; i < len(result)/2; i++ {
			j := len(result) - 1 - i
			result[i], result[j] = result[j], result[i]
		}
	}

	return result, true

}

// New List Operations
func ListPush(key string, elements []string, left bool) int {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	value, exists := data[key]
	//for very first value, to create one
	if !exists {
		value = &RedisValue{
			Type: LIST,
			List: make([]string, 0),
		}
		data[key] = value
	}

	//type check
	if value.Type != LIST {
		return -1
	}

	//expiry check
	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		//reset expired list
		value.List = make([]string, 0)
		value.Expiry = nil
	}

	// add elements
	if left {
		// LPUSH: prepend elements (reverse order for multiples)
		for i := len(elements) - 1; i >= 0; i-- {
			value.List = append([]string{elements[i]}, value.List...)
		}
	} else {
		//RPUSH: append elements
		value.List = append(value.List, elements...)
	}
	length := len(value.List)

	dataMutex.Unlock()

	go NotifyBlockingClients(key)

	dataMutex.Lock()
	return length
}

func ListBlockingPopImmediate(keys []string, left bool) (string, string, bool) {

	dataMutex.Lock()
	defer dataMutex.Unlock()

	for _, key := range keys {
		value, exists := data[key]
		if !exists {
			continue
		}

		if value.Type != LIST {
			continue
		}

		if value.Expiry != nil && time.Now().After(*value.Expiry) {
			delete(data, key)
			continue
		}
		listlen := len(value.List)

		if listlen == 0 {
			continue
		}

		var element string
		if left {
			element = value.List[0]
			value.List = value.List[1:]
		} else {
			element = value.List[listlen-1]
			value.List = value.List[:listlen-1]
		}

		return key, element, true
	}
	return "", "", false
}

// Registering a client to wait for elements on keys
func RegisterBlockingClient(keys []string, left bool, timeout time.Duration) *BlockingClient {
	blockingMutex.Lock()
	defer blockingMutex.Unlock()

	client := &BlockingClient{
		Keys:     keys,
		Left:     left,
		Response: make(chan BlockingResult, 1),
		Timeout:  timeout,
	}

	// Register client for each key
	for _, key := range keys {
		blockingClients[key] = append(blockingClients[key], client)
	}
	return client
}

func UnregisterBlockingClient(client *BlockingClient) {
	blockingMutex.Lock()
	defer blockingMutex.Unlock()

	for _, key := range client.Keys {
		clients := blockingClients[key]
		for i, c := range clients {
			if c == client {
				blockingClients[key] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		// Clean Up empty key entries
		if len(blockingClients[key]) == 0 {
			delete(blockingClients, key)
		}
	}
}

// Notify waiting clients when new elements are added
func NotifyBlockingClients(key string) {
	blockingMutex.Lock()
	clients := blockingClients[key]
	// Make a copy to avoid holding the lock too long

	clientsCopy := make([]*BlockingClient, len(clients))
	copy(clientsCopy, clients)
	blockingMutex.Unlock()

	if len(clientsCopy) == 0 {
		return
	}

	for _, client := range clientsCopy {
		foundKey, element, found := ListBlockingPopImmediate(client.Keys, client.Left)
		if !found {
			break
		}

		select {
		case client.Response <- BlockingResult{Key: foundKey, Value: element, Success: true}:
			// Successfully notified client, remove from waiting list
			UnregisterBlockingClient(client)
		default:
			// Client channel is full or closed, remove it
			UnregisterBlockingClient(client)
		}
	}
}
