package store

import (
	"sync"
	"time"
)

var (
	data      = make(map[string]string)
	expiries  = make(map[string]time.Time)
	dataMutex sync.RWMutex
)

func Set(key, val string, ttl time.Duration) {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	data[key] = val
	if ttl > 0 {
		expiries[key] = time.Now().Add(ttl)
	} else {
		delete(expiries, key)
	}

}

func Get(key string) (string, bool) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	if expiry, ok := expiries[key]; ok {
		if time.Now().After(expiry) {
			delete(data, key)
			delete(expiries, key)
			return "", false
		}
	}

	val, ok := data[key]
	return val, ok

}
