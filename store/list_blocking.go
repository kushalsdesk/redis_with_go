package store

import (
	"sync"
	"time"
)

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
