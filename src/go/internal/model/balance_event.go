package model

import "time"

// BalanceEvent stores raw balance change events received from RabbitMQ.
// It's used for idempotency (unique EventID) and history.
type BalanceEvent struct {
	ID        uint      `gorm:"primaryKey"`
	EventID   string    `gorm:"type:char(36);uniqueIndex;not null"`
	UserID    uint      `gorm:"not null"`
	WalletID  uint      `gorm:"not null"`
	Balance   int64     `gorm:"not null"`
	UpdatedAt time.Time `gorm:"index"`
	CreatedAt time.Time `gorm:"not null"`
}
