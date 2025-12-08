package service

import "time"

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

type Handler interface {
	Handle(Event) error
}
