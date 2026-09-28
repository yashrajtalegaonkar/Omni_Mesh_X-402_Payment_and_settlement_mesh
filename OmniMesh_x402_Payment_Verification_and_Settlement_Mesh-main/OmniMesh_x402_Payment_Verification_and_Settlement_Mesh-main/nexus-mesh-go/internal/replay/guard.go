package replay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

type cacheEntry struct {
	timestamp time.Time
	status    string
	traceID   string
}

type ReplayGuard struct {
	ttl   time.Duration
	cache map[string]cacheEntry
	mu    sync.RWMutex
}

func NewReplayGuard(ttlSeconds int) *ReplayGuard {
	rg := &ReplayGuard{
		ttl:   time.Duration(ttlSeconds) * time.Second,
		cache: make(map[string]cacheEntry),
	}
	go rg.startCleanupLoop()
	return rg
}

func (rg *ReplayGuard) ComputeHash(payload map[string]interface{}) string {
	data, _ := json.Marshal(payload)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (rg *ReplayGuard) IsDuplicate(payloadHash string) (bool, string) {
	rg.mu.RLock()
	defer rg.mu.RUnlock()

	entry, exists := rg.cache[payloadHash]
	if exists {
		return true, entry.traceID
	}
	return false, ""
}

func (rg *ReplayGuard) Register(payloadHash, traceID string, status ...string) {
	st := "REGISTERED"
	if len(status) > 0 {
		st = status[0]
	}
	rg.mu.Lock()
	defer rg.mu.Unlock()

	rg.cache[payloadHash] = cacheEntry{
		timestamp: time.Now(),
		status:    st,
		traceID:   traceID,
	}
}

func (rg *ReplayGuard) UpdateStatus(payloadHash, newStatus string) {
	rg.mu.Lock()
	defer rg.mu.Unlock()

	if entry, exists := rg.cache[payloadHash]; exists {
		entry.status = newStatus
		rg.cache[payloadHash] = entry
	}
}

func (rg *ReplayGuard) startCleanupLoop() {
	ticker := time.NewTicker(60 * time.Second)
	for range ticker.C {
		rg.mu.Lock()
		now := time.Now()
		for hash, entry := range rg.cache {
			if now.Sub(entry.timestamp) > rg.ttl {
				delete(rg.cache, hash)
			}
		}
		rg.mu.Unlock()
	}
}

var DefaultGuard = NewReplayGuard(300)
