package store

import "time"

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
