package store

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func StreamAdd(key, id string, fields map[string]string) (string, error) {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	value, exists := data[key]
	if !exists {
		//new stream
		value = &RedisValue{
			Type: STREAM,
			Stream: &Stream{
				Entries: make([]StreamEntry, 0),
				LastID:  "",
			},
		}
		data[key] = value
	}
	if value.Type != STREAM {
		return "", fmt.Errorf("WRONGTYPE Operation against a key holding wrong kind of value")
	}

	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		//resetting expired stream
		value.Stream = &Stream{
			Entries: make([]StreamEntry, 0),
			LastID:  "",
		}
		value.Expiry = nil
	}

	var finalID string
	var err error

	switch {
	case id == "*":
		finalID = GenerateAutoID(value.Stream.LastID)

	case strings.HasSuffix(id, "-*"):
		timestampStr := strings.TrimSuffix(id, "-*")
		timestamp, parseErr := strconv.ParseInt(timestampStr, 10, 64)
		if parseErr != nil {
			return "", fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
		}

		if timestamp < 0 {
			return "", fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
		}

		sequence, seqErr := GenerateNextSequence(timestamp, value.Stream.LastID)
		if seqErr != nil {
			return "", seqErr
		}

		finalID = fmt.Sprintf("%d-%d", timestamp, sequence)

	default:
		err = ValidateStreamID(id, value.Stream.LastID)
		if err != nil {
			return "", err
		}
		finalID = id
	}

	entry := StreamEntry{
		ID:     finalID,
		Fields: fields,
	}
	value.Stream.Entries = append(value.Stream.Entries, entry)
	value.Stream.LastID = finalID

	go NotifyStreamBlockingClients(key)

	return finalID, nil

}

func ParseStreamID(idStr string) (*StreamID, error) {
	if idStr == "" {
		return nil, fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}

	parts := strings.Split(idStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}

	sequence, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}

	//negative checking
	if timestamp < 0 {
		return nil, fmt.Errorf("ERR the ID specified in XADD must be greater than 0-0")
	}

	if sequence < 0 {
		return nil, fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}

	return &StreamID{
		Timestamp: timestamp,
		Sequence:  sequence,
	}, nil
}

// to convert StreamID back to string
func (s *StreamID) String() string {
	return fmt.Sprintf("%d-%d", s.Timestamp, s.Sequence)
}

func CompareStreamIDs(id1, id2 string) int {
	parsed1, err1 := ParseStreamID(id1)
	parsed2, err2 := ParseStreamID(id2)

	if err1 != nil || err2 != nil {
		return strings.Compare(id1, id2)
	}

	if parsed1.Timestamp < parsed2.Timestamp {
		return -1
	}

	if parsed1.Timestamp > parsed2.Timestamp {
		return 1
	}

	if parsed1.Sequence < parsed2.Sequence {
		return -1
	}

	if parsed1.Sequence > parsed2.Sequence {
		return 1
	}
	return 0
}

func ValidateStreamID(newID, lastID string) error {
	parsed, err := ParseStreamID(newID)
	if err != nil {
		return err
	}

	if parsed.Timestamp == 0 && parsed.Sequence == 0 {
		return fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")

	}
	if lastID == "" {
		return nil
	}

	comparison := CompareStreamIDs(newID, lastID)
	if comparison <= 0 {
		return fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}
	return nil
}

func GenerateNextSequence(timestamp int64, lastID string) (int64, error) {
	if lastID == "" {
		return 0, nil
	}

	lastParsed, err := ParseStreamID(lastID)
	if err != nil {
		return 0, err
	}

	if timestamp > lastParsed.Timestamp {
		return 0, nil
	}

	if timestamp == lastParsed.Timestamp {
		return lastParsed.Sequence + 1, nil
	}

	return 0, fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
}

func GenerateAutoID(lastID string) string {
	currentTime := time.Now().UnixMilli()

	if lastID == "" {
		return fmt.Sprintf("%d-0", currentTime)
	}
	lastParsed, err := ParseStreamID(lastID)
	if err != nil {
		return fmt.Sprintf("%d-0", currentTime)
	}

	if currentTime > lastParsed.Timestamp {
		return fmt.Sprintf("%d-0", currentTime)
	}

	if currentTime == lastParsed.Timestamp {
		return fmt.Sprintf("%d-%d", currentTime, lastParsed.Sequence+1)
	}

	return fmt.Sprintf("%d-0", lastParsed.Timestamp+1)

}

// StreaRange returns entries within specified ID range
func StreamRange(key, start, end string) ([]StreamEntry, error) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	value, exists := data[key]
	if !exists {
		return []StreamEntry{}, nil
	}

	if value.Type != STREAM {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		return []StreamEntry{}, nil
	}

	stream := value.Stream
	if len(stream.Entries) == 0 {
		return []StreamEntry{}, nil
	}

	startID, endID := resolveRangeIDs(start, end, stream)

	var result []StreamEntry
	for _, entry := range stream.Entries {
		entryID := entry.ID

		if CompareStreamIDs(entryID, startID) >= 0 && CompareStreamIDs(entryID, endID) <= 0 {
			result = append(result, entry)
		}
	}

	return result, nil

}

func resolveRangeIDs(start, end string, stream *Stream) (string, string) {
	startID := start
	endID := end

	//handling minimum possible ID/start of stream
	if start == "-" {
		if len(stream.Entries) > 0 {
			startID = stream.Entries[0].ID
		} else {
			startID = "0-1" // minimum valid id
		}
	}

	//handling maximum possible ID/end of stream
	if end == "+" {
		if len(stream.Entries) > 0 {
			endID = stream.Entries[len(stream.Entries)-1].ID
		} else {
			endID = "9223372036854775807-9223372036854775807"
		}
	}

	return startID, endID
}

// StreamReadFrom returns entries after the given ID
func StreamReadFrom(key, startID string, count int) ([]StreamEntry, error) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	value, exists := data[key]
	if !exists {
		return []StreamEntry{}, nil
	}

	if value.Type != STREAM {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		return []StreamEntry{}, nil
	}

	stream := value.Stream
	if len(stream.Entries) == 0 {
		return []StreamEntry{}, nil
	}

	var result []StreamEntry
	for _, entry := range stream.Entries {
		if CompareStreamIDs(entry.ID, startID) > 0 {
			result = append(result, entry)

			if count > 0 && len(result) >= count {
				break
			}
		}
	}
	return result, nil
}

func GetStreamLastID(key string) string {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	value, exists := data[key]
	if !exists {
		return ""
	}

	if value.Type != STREAM {
		return ""
	}

	if value.Expiry != nil && time.Now().After(*value.Expiry) {
		return ""
	}
	return value.Stream.LastID
}

func StreamReadFromImmediate(streamKeys, startIDs []string, count int) ([]StreamReadResult, bool) {
	results := make([]StreamReadResult, 0)

	for i, key := range streamKeys {
		startID := startIDs[i]

		if startID == "$" {
			lastID := GetStreamLastID(key)
			if lastID != "" {
				startID = lastID
			} else {
				continue
			}
		}

		entries, err := StreamReadFrom(key, startID, count)
		if err != nil {
			continue
		}

		if len(entries) > 0 {
			results = append(results, StreamReadResult{
				StreamKey: key,
				Entries:   entries,
			})
		}
	}

	return results, len(results) > 0
}
