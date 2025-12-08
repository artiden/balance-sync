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

// A simple in-memory cache of balances.
type Cache struct {
	m sync.Map // key: CacheKey, value: CacheValue
}

func NewCache() *Cache { return &Cache{} }

// SetIfNewer updates the cache only if the given timestamp is newer than
// what is already stored for the same key. If there is no existing value,
// it is always inserted.
func (c *Cache) SetIfNewer(userID, walletID uint, balance int64, updatedAt time.Time) {
	key := CacheKey{UserID: userID, WalletID: walletID}
	newVal := CacheValue{Balance: balance, UpdatedAt: updatedAt}

	if cur, ok := c.m.Load(key); ok {
		current := cur.(CacheValue)
		if !updatedAt.After(current.UpdatedAt) {
			// Ignore
			return
		}
	}
	c.m.Store(key, newVal)
}

// Snapshot returns a copy of the current cache contents.
func (c *Cache) Snapshot() map[CacheKey]CacheValue {
	out := make(map[CacheKey]CacheValue)
	c.m.Range(func(k, v interface{}) bool {
		out[k.(CacheKey)] = v.(CacheValue)
		return true
	})
	return out
}

// See README: Cache Logic Note. We should implement additional functionality to use it in production
