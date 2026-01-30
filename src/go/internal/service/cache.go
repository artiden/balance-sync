package service

import (
	"sync"
	"time"
)

// CacheKey identifies a particular wallet of a user.
type CacheKey struct {
	UserID   uint
	WalletID uint
}

// CacheValue stores the last known balance.
type CacheValue struct {
	Balance   int64
	UpdatedAt time.Time
}

// A simple in-memory cache of balances with per-key locking to prevent race conditions.
type Cache struct {
	m   sync.Map // key: CacheKey, value: CacheValue
	mu  sync.Map // key: CacheKey, value: *sync.Mutex (per-key locks)
}

func NewCache() *Cache { return &Cache{} }

// getMutex returns the mutex for a given key, creating one if necessary.
func (c *Cache) getMutex(key CacheKey) *sync.Mutex {
	mu, _ := c.mu.LoadOrStore(key, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// SetIfNewer updates the cache only if the given timestamp is newer than
// what is already stored for the same key. If there is no existing value,
// it is always inserted. Uses per-key locking to prevent race conditions.
func (c *Cache) SetIfNewer(userID, walletID uint, balance int64, updatedAt time.Time) {
	key := CacheKey{UserID: userID, WalletID: walletID}
	newVal := CacheValue{Balance: balance, UpdatedAt: updatedAt}

	// Lock per-key to prevent race condition between Load and Store
	mu := c.getMutex(key)
	mu.Lock()
	defer mu.Unlock()

	if cur, ok := c.m.Load(key); ok {
		current, ok := cur.(CacheValue)
		if !ok {
			// Unexpected type - overwrite
			c.m.Store(key, newVal)
			return
		}
		if !updatedAt.After(current.UpdatedAt) {
			// Ignore stale update
			return
		}
	}
	c.m.Store(key, newVal)
}

// Snapshot returns a copy of the current cache contents.
func (c *Cache) Snapshot() map[CacheKey]CacheValue {
	out := make(map[CacheKey]CacheValue)
	c.m.Range(func(k, v interface{}) bool {
		key, ok := k.(CacheKey)
		if !ok {
			return true
		}
		val, ok := v.(CacheValue)
		if !ok {
			return true
		}
		out[key] = val
		return true
	})
	return out
}

// See README: Cache Logic Note. We should implement additional functionality to use it in production
