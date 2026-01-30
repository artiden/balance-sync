package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepository implements Repository interface for testing
type mockRepository struct {
	applyEventFn    func(Event) (bool, error)
	latestStatesFn  func() ([]LatestBalance, error)
	appliedEvents   []Event
}

func (m *mockRepository) ApplyEvent(e Event) (bool, error) {
	m.appliedEvents = append(m.appliedEvents, e)
	if m.applyEventFn != nil {
		return m.applyEventFn(e)
	}
	return true, nil
}

func (m *mockRepository) LatestStates() ([]LatestBalance, error) {
	if m.latestStatesFn != nil {
		return m.latestStatesFn()
	}
	return nil, nil
}

func TestHandler_Handle_AppliesEventToRepository(t *testing.T) {
	cache := NewCache()
	repo := &mockRepository{}
	h := NewHandler(cache, repo)

	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	err := h.Handle(event)

	require.NoError(t, err)
	assert.Len(t, repo.appliedEvents, 1)
	assert.Equal(t, event.EventID, repo.appliedEvents[0].EventID)
}

func TestHandler_Handle_UpdatesCacheOnSuccess(t *testing.T) {
	cache := NewCache()
	repo := &mockRepository{
		applyEventFn: func(e Event) (bool, error) {
			return true, nil // Event was applied
		},
	}
	h := NewHandler(cache, repo)

	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	err := h.Handle(event)

	require.NoError(t, err)

	// Verify cache was updated
	snapshot := cache.Snapshot()
	key := CacheKey{UserID: 1, WalletID: 1}
	assert.Equal(t, int64(1000), snapshot[key].Balance)
}

func TestHandler_Handle_DoesNotUpdateCacheForDuplicate(t *testing.T) {
	cache := NewCache()
	repo := &mockRepository{
		applyEventFn: func(e Event) (bool, error) {
			return false, nil // Event was duplicate/stale
		},
	}
	h := NewHandler(cache, repo)

	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	err := h.Handle(event)

	require.NoError(t, err)

	// Cache should be empty
	snapshot := cache.Snapshot()
	assert.Empty(t, snapshot)
}

func TestHandler_Handle_ReturnsErrorOnRepositoryFailure(t *testing.T) {
	cache := NewCache()
	expectedErr := errors.New("database error")
	repo := &mockRepository{
		applyEventFn: func(e Event) (bool, error) {
			return false, expectedErr
		},
	}
	h := NewHandler(cache, repo)

	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	err := h.Handle(event)

	assert.ErrorIs(t, err, expectedErr)

	// Cache should not be updated on error
	snapshot := cache.Snapshot()
	assert.Empty(t, snapshot)
}

func TestHandler_Handle_ProcessesMultipleEventsSequentially(t *testing.T) {
	cache := NewCache()
	repo := &mockRepository{}
	h := NewHandler(cache, repo)

	now := time.Now().UTC()
	events := []Event{
		{EventID: uuid.New().String(), UserID: 1, WalletID: 1, WalletBalance: 100, UpdatedAt: now},
		{EventID: uuid.New().String(), UserID: 1, WalletID: 1, WalletBalance: 200, UpdatedAt: now.Add(time.Second)},
		{EventID: uuid.New().String(), UserID: 1, WalletID: 1, WalletBalance: 300, UpdatedAt: now.Add(2 * time.Second)},
	}

	for _, event := range events {
		err := h.Handle(event)
		require.NoError(t, err)
	}

	assert.Len(t, repo.appliedEvents, 3)

	// Cache should have the latest value
	snapshot := cache.Snapshot()
	key := CacheKey{UserID: 1, WalletID: 1}
	assert.Equal(t, int64(300), snapshot[key].Balance)
}
