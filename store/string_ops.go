package store

import (
	"fmt"
	"math"
	"strconv"
	"time"
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

	if isExpired(value, key) {
		return "", false
	}

	if value.Type != STRING {
		return "", false
	}

	return value.String, true
}

func Delete(key string) bool {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	_, exists := data[key]
	if exists {
		delete(data, key)
		return true
	}
	return false
}

// Increment operations for replication
func Increment(key string) (int64, error) {
	return IncrementBy(key, 1)
}

func Decrement(key string) (int64, error) {
	return IncrementBy(key, -1)
}

func IncrementBy(key string, amount int64) (int64, error) {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	value, exists := data[key]
	var currentVal int64 = 0

	if exists {
		if value.Type != STRING {
			return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}

		if value.Expiry != nil && time.Now().After(*value.Expiry) {
			// Key expired, treat as non-existent
			currentVal = 0
		} else {
			parsedVal, err := strconv.ParseInt(value.String, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("value is not an integer or out of range")
			}
			currentVal = parsedVal
		}
	}

	// Check for overflow/underflow
	if amount > 0 {
		if currentVal > 0 && amount > math.MaxInt64-currentVal {
			return 0, fmt.Errorf("increment or decrement would overflow")
		}
	} else {
		if currentVal < 0 && amount < math.MinInt64-currentVal {
			return 0, fmt.Errorf("increment or decrement would overflow")
		}
	}

	newValue := currentVal + amount

	// Store the new value
	if !exists {
		data[key] = &RedisValue{
			Type:   STRING,
			String: strconv.FormatInt(newValue, 10),
		}
	} else {
		value.String = strconv.FormatInt(newValue, 10)
		// Clear expiry when modifying
		value.Expiry = nil
	}

	return newValue, nil
}

func DecrementBy(key string, amount int64) (int64, error) {
	return IncrementBy(key, -amount)
}
