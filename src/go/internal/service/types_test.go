package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEvent_Validate_ValidEvent(t *testing.T) {
	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	err := event.Validate()

	assert.NoError(t, err)
}

func TestEvent_Validate_EmptyEventID(t *testing.T) {
	event := Event{
		EventID:       "",
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	err := event.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "eventId")
}

func TestEvent_Validate_ZeroUserID(t *testing.T) {
	event := Event{
		EventID:       uuid.New().String(),
		UserID:        0,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	err := event.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userId")
}

func TestEvent_Validate_ZeroWalletID(t *testing.T) {
	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      0,
		WalletBalance: 1000,
		UpdatedAt:     time.Now().UTC(),
	}

	err := event.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "walletId")
}

func TestEvent_Validate_ZeroUpdatedAt(t *testing.T) {
	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 1000,
		UpdatedAt:     time.Time{}, // Zero time
	}

	err := event.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updatedAt")
}

func TestEvent_Validate_NegativeBalance(t *testing.T) {
	// Negative balance is allowed by design
	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: -500,
		UpdatedAt:     time.Now().UTC(),
	}

	err := event.Validate()

	assert.NoError(t, err, "negative balance should be allowed")
}

func TestEvent_Validate_ZeroBalance(t *testing.T) {
	event := Event{
		EventID:       uuid.New().String(),
		UserID:        1,
		WalletID:      1,
		WalletBalance: 0,
		UpdatedAt:     time.Now().UTC(),
	}

	err := event.Validate()

	assert.NoError(t, err, "zero balance should be allowed")
}
