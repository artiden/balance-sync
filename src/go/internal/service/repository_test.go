package service

import (
	"testing"
	"time"

	"goservice/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing.
// Note: Some MySQL-specific behaviors (like error code 1062) won't be tested here.
// For full integration tests, use the actual MySQL database.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.Balance{}, &model.BalanceEvent{})
	require.NoError(t, err)

	return db
}

func TestRepository_ApplyEvent_CreatesNewBalance(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	updated, err := repo.ApplyEvent(event)

	require.NoError(t, err)
	assert.True(t, updated, "should return true for new balance")

	// Verify balance was created
	var balance model.Balance
	err = db.Where("user_id = ? AND wallet_id = ?", event.UserID, event.WalletID).First(&balance).Error
	require.NoError(t, err)
	assert.Equal(t, event.WalletBalance, balance.Balance)
	assert.Equal(t, event.UserID, balance.UserID)
	assert.Equal(t, event.WalletID, balance.WalletID)
}

func TestRepository_ApplyEvent_UpdatesExistingBalance(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Create initial balance
	initialTime := time.Now().UTC().Add(-time.Hour)
	initialEvent := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     initialTime,
	}
	_, err := repo.ApplyEvent(initialEvent)
	require.NoError(t, err)

	// Update with newer event
	newTime := time.Now().UTC()
	updateEvent := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 2000,
		UpdatedAt:     newTime,
	}

	updated, err := repo.ApplyEvent(updateEvent)

	require.NoError(t, err)
	assert.True(t, updated, "should return true for updated balance")

	// Verify balance was updated
	var balance model.Balance
	err = db.Where("user_id = ? AND wallet_id = ?", updateEvent.UserID, updateEvent.WalletID).First(&balance).Error
	require.NoError(t, err)
	assert.Equal(t, int64(2000), balance.Balance)
}

func TestRepository_ApplyEvent_IgnoresStaleEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Create initial balance with recent timestamp
	recentTime := time.Now().UTC()
	initialEvent := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 2000,
		UpdatedAt:     recentTime,
	}
	_, err := repo.ApplyEvent(initialEvent)
	require.NoError(t, err)

	// Try to apply older event
	oldTime := recentTime.Add(-time.Hour)
	staleEvent := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     oldTime,
	}

	updated, err := repo.ApplyEvent(staleEvent)

	require.NoError(t, err)
	assert.False(t, updated, "should return false for stale event")

	// Verify balance was NOT updated
	var balance model.Balance
	err = db.Where("user_id = ? AND wallet_id = ?", 1, 1).First(&balance).Error
	require.NoError(t, err)
	assert.Equal(t, int64(2000), balance.Balance, "balance should remain unchanged")
}

func TestRepository_ApplyEvent_StoresBalanceEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	eventID := uuid.New().String()
	event := Event{
		EventID:       eventID,
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	_, err := repo.ApplyEvent(event)
	require.NoError(t, err)

	// Verify event was stored
	var balanceEvent model.BalanceEvent
	err = db.Where("event_id = ?", eventID).First(&balanceEvent).Error
	require.NoError(t, err)
	assert.Equal(t, eventID, balanceEvent.EventID)
	assert.Equal(t, event.UserID, balanceEvent.UserID)
	assert.Equal(t, event.WalletID, balanceEvent.WalletID)
	assert.Equal(t, event.WalletBalance, balanceEvent.Balance)
}

func TestRepository_ApplyEvent_HandlesMultipleUsers(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	now := time.Now().UTC()

	// Create events for multiple users
	events := []Event{
		{EventID: uuid.New().String(), UserID: 1, WalletID: 1, WalletBalance: 100, UpdatedAt: now},
		{EventID: uuid.New().String(), UserID: 2, WalletID: 2, WalletBalance: 200, UpdatedAt: now},
		{EventID: uuid.New().String(), UserID: 3, WalletID: 3, WalletBalance: 300, UpdatedAt: now},
	}

	for _, event := range events {
		updated, err := repo.ApplyEvent(event)
		require.NoError(t, err)
		assert.True(t, updated)
	}

	// Verify all balances
	var balances []model.Balance
	err := db.Find(&balances).Error
	require.NoError(t, err)
	assert.Len(t, balances, 3)
}

func TestRepository_LatestStates_ReturnsAllBalances(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	now := time.Now().UTC()

	// Create some balances
	events := []Event{
		{EventID: uuid.New().String(), UserID: 1, WalletID: 1, WalletBalance: 100, UpdatedAt: now},
		{EventID: uuid.New().String(), UserID: 2, WalletID: 2, WalletBalance: 200, UpdatedAt: now},
	}

	for _, event := range events {
		_, err := repo.ApplyEvent(event)
		require.NoError(t, err)
	}

	states, err := repo.LatestStates()

	require.NoError(t, err)
	assert.Len(t, states, 2)

	// Verify states contain correct data
	stateMap := make(map[uint]LatestBalance)
	for _, s := range states {
		stateMap[s.UserID] = s
	}

	assert.Equal(t, int64(100), stateMap[1].Balance)
	assert.Equal(t, int64(200), stateMap[2].Balance)
}

func TestRepository_LatestStates_ReturnsEmptyForNoData(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	states, err := repo.LatestStates()

	require.NoError(t, err)
	assert.Empty(t, states)
}
