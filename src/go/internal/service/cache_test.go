package service

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetIfNewer_InsertsNewValue(t *testing.T) {
	cache := NewCache()
	now := time.Now().UTC()

	cache.SetIfNewer(1, 1, 1000, now)

	snapshot := cache.Snapshot()
	assert.Len(t, snapshot, 1)

	key := CacheKey{UserID: 1, WalletID: 1}
	value, exists := snapshot[key]
	assert.True(t, exists)
	assert.Equal(t, int64(1000), value.Balance)
	assert.Equal(t, now, value.UpdatedAt)
}

func TestCache_SetIfNewer_UpdatesWithNewerTimestamp(t *testing.T) {
	cache := NewCache()
	oldTime := time.Now().UTC().Add(-time.Hour)
	newTime := time.Now().UTC()

	// Insert initial value
	cache.SetIfNewer(1, 1, 1000, oldTime)

	// Update with newer timestamp
	cache.SetIfNewer(1, 1, 2000, newTime)

	snapshot := cache.Snapshot()
	key := CacheKey{UserID: 1, WalletID: 1}
	assert.Equal(t, int64(2000), snapshot[key].Balance)
	assert.Equal(t, newTime, snapshot[key].UpdatedAt)
}

func TestCache_SetIfNewer_IgnoresOlderTimestamp(t *testing.T) {
	cache := NewCache()
	oldTime := time.Now().UTC().Add(-time.Hour)
	newTime := time.Now().UTC()

	// Insert with newer timestamp first
	cache.SetIfNewer(1, 1, 2000, newTime)

	// Try to update with older timestamp
	cache.SetIfNewer(1, 1, 1000, oldTime)

	snapshot := cache.Snapshot()
	key := CacheKey{UserID: 1, WalletID: 1}
	assert.Equal(t, int64(2000), snapshot[key].Balance, "balance should not change for older event")
	assert.Equal(t, newTime, snapshot[key].UpdatedAt)
}

func TestCache_SetIfNewer_IgnoresSameTimestamp(t *testing.T) {
	cache := NewCache()
	sameTime := time.Now().UTC()

	cache.SetIfNewer(1, 1, 1000, sameTime)
	cache.SetIfNewer(1, 1, 2000, sameTime) // Same timestamp

	snapshot := cache.Snapshot()
	key := CacheKey{UserID: 1, WalletID: 1}
	// First value should win when timestamps are equal
	assert.Equal(t, int64(1000), snapshot[key].Balance)
}

func TestCache_SetIfNewer_HandlesMultipleKeys(t *testing.T) {
	cache := NewCache()
	now := time.Now().UTC()

	cache.SetIfNewer(1, 1, 100, now)
	cache.SetIfNewer(2, 2, 200, now)
	cache.SetIfNewer(3, 3, 300, now)

	snapshot := cache.Snapshot()
	assert.Len(t, snapshot, 3)

	assert.Equal(t, int64(100), snapshot[CacheKey{UserID: 1, WalletID: 1}].Balance)
	assert.Equal(t, int64(200), snapshot[CacheKey{UserID: 2, WalletID: 2}].Balance)
	assert.Equal(t, int64(300), snapshot[CacheKey{UserID: 3, WalletID: 3}].Balance)
}

func TestCache_SetIfNewer_ConcurrentAccess(t *testing.T) {
	cache := NewCache()
	baseTime := time.Now().UTC()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes to the same key
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Each goroutine writes with progressively newer timestamps
			timestamp := baseTime.Add(time.Duration(idx) * time.Millisecond)
			cache.SetIfNewer(1, 1, int64(idx*100), timestamp)
		}(i)
	}

	wg.Wait()

	snapshot := cache.Snapshot()
	key := CacheKey{UserID: 1, WalletID: 1}

	// The value should be from the newest timestamp (idx = 99)
	assert.Equal(t, int64(9900), snapshot[key].Balance)
}

func TestCache_SetIfNewer_ConcurrentAccessDifferentKeys(t *testing.T) {
	cache := NewCache()
	now := time.Now().UTC()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes to different keys
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cache.SetIfNewer(uint(idx), uint(idx), int64(idx*100), now)
		}(i)
	}

	wg.Wait()

	snapshot := cache.Snapshot()
	assert.Len(t, snapshot, numGoroutines)
}

func TestCache_Snapshot_ReturnsEmptyForNewCache(t *testing.T) {
	cache := NewCache()

	snapshot := cache.Snapshot()

	assert.Empty(t, snapshot)
}

func TestCache_Snapshot_ReturnsCopy(t *testing.T) {
	cache := NewCache()
	now := time.Now().UTC()
	cache.SetIfNewer(1, 1, 1000, now)

	snapshot1 := cache.Snapshot()
	snapshot2 := cache.Snapshot()

	// Modifying one snapshot should not affect the other
	delete(snapshot1, CacheKey{UserID: 1, WalletID: 1})

	assert.Empty(t, snapshot1)
	assert.Len(t, snapshot2, 1)
}
