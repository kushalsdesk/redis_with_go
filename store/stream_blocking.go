package store

import (
	"sync"
	"time"
)

// StreamBlockingClient represents a client waiting for stream entries
type StreamBlockingClient struct {
	StreamKeys []string
	StartIDs   []string
	Count      int
	Response   chan StreamBlockingResult
	Timeout    time.Duration
}

type StreamBlockingResult struct {
	Results []StreamReadResult
	Success bool
	Timeout bool
}

type StreamReadResult struct {
	StreamKey string
	Entries   []StreamEntry
}

var (
	streamBlockingClients = make(map[string][]*StreamBlockingClient)
	streamBlockingMutex   sync.Mutex
)

// RegisterStreamBlockingClient registers a client to wait for stream entries

func RegisterStreamBlockingClient(streamKeys, startIDs []string, count int, timeout time.Duration) *StreamBlockingClient {
	streamBlockingMutex.Lock()
	defer streamBlockingMutex.Unlock()

	processedStartIDs := make([]string, len(startIDs))
	for i, startID := range startIDs {
		if startID == "$" {
			lastID := GetStreamLastID(streamKeys[i])
			if lastID != "" {
				processedStartIDs[i] = lastID
			} else {
				processedStartIDs[i] = "0-0"
			}
		} else {
			processedStartIDs[i] = startID
		}
	}

	client := &StreamBlockingClient{
		StreamKeys: streamKeys,
		StartIDs:   processedStartIDs,
		Count:      count,
		Response:   make(chan StreamBlockingResult, 1),
		Timeout:    timeout,
	}

	//Register client for each stream key

	for _, key := range streamKeys {
		streamBlockingClients[key] = append(streamBlockingClients[key], client)
	}
	return client
}

// UnregisterStreamBlockingClient removes a waiting client
func UnregisterStreamBlockingClient(client *StreamBlockingClient) {
	streamBlockingMutex.Lock()
	defer streamBlockingMutex.Unlock()

	for _, key := range client.StreamKeys {
		clients := streamBlockingClients[key]
		for i, c := range clients {
			if c == client {
				streamBlockingClients[key] = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		if len(streamBlockingClients[key]) == 0 {
			delete(streamBlockingClients, key)
		}
	}
}

func NotifyStreamBlockingClients(key string) {
	streamBlockingMutex.Lock()
	clients := streamBlockingClients[key]

	//making a copy to avoid holding the lock too long
	clientsCopy := make([]*StreamBlockingClient, len(clients))
	copy(clientsCopy, clients)
	streamBlockingMutex.Unlock()

	if len(clientsCopy) == 0 {
		return
	}

	for _, client := range clientsCopy {
		keyIndex := -1
		for i, streamKey := range client.StreamKeys {
			if streamKey == key {
				keyIndex = i
				break
			}
		}
		if keyIndex == -1 {
			continue
		}

		startID := client.StartIDs[keyIndex]

		// if startID == "$" {
		// 	lastID := GetStreamLastID(key)
		// 	if lastID != "" {
		// 		startID = lastID
		// 	} else {
		// 		continue
		// 	}
		// }

		entries, err := StreamReadFrom(key, startID, client.Count)

		if err != nil || len(entries) == 0 {
			continue
		}

		results := []StreamReadResult{
			{
				StreamKey: key,
				Entries:   entries,
			},
		}

		select {
		case client.Response <- StreamBlockingResult{
			Results: results,
			Success: true,
			Timeout: false,
		}:
			UnregisterStreamBlockingClient(client)
		default:
			UnregisterStreamBlockingClient(client)
		}
	}
}
