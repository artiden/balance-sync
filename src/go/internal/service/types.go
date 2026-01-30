package service

import (
	"errors"
	"time"
)

// Event describes a balance update coming from RabbitMQ.
type Event struct {
	EventID string `json:"eventId"`
	// Probably we'll need it in the future...
	EventType     string    `json:"eventType"`
	UserID        uint      `json:"userId"`
	WalletID      uint      `json:"walletId"`
	WalletBalance int64     `json:"walletBalance"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Validate checks that the event has all required fields.
func (e Event) Validate() error {
	if e.EventID == "" {
		return errors.New("eventId is required")
	}
	if e.UserID == 0 {
		return errors.New("userId is required")
	}
	if e.WalletID == 0 {
		return errors.New("walletId is required")
	}
	if e.UpdatedAt.IsZero() {
		return errors.New("updatedAt is required")
	}
	return nil
}

type Handler interface {
	Handle(Event) error
}
