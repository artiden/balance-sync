package model

import "time"

// Balance represents the latest known balance for a particular wallet of a user.
type Balance struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index:idx_user_wallet,unique"`
	WalletID  uint      `gorm:"index:idx_user_wallet,unique"`
	Balance   int64     `gorm:"not null"`
	UpdatedAt time.Time `gorm:"index"`
	CreatedAt time.Time `gorm:"not null"`
}
