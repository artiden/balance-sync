package service

import (
	"log"
	"time"

	"goservice/internal/model"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// Represents the latest known balance
type LatestBalance struct {
	UserID    uint
	WalletID  uint
	Balance   int64
	UpdatedAt time.Time
}

// Repository abstracts persistence for balances & events.
type Repository interface {
	ApplyEvent(event Event) (bool, error)

	LatestStates() ([]LatestBalance, error)
}

type repo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository {
	return &repo{db: db}
}

func (r *repo) ApplyEvent(e Event) (bool, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return false, tx.Error
	}

	hist := &model.BalanceEvent{
		EventID:   e.EventID,
		UserID:    e.UserID,
		WalletID:  e.WalletID,
		Balance:   e.WalletBalance,
		UpdatedAt: e.UpdatedAt,
	}

	if err := tx.Create(hist).Error; err != nil {
		// Event already processed.
		if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1062 {
			log.Printf("repository: duplicate event %s – ignored", e.EventID)
			if rbErr := tx.Rollback().Error; rbErr != nil {
				log.Printf("repository: rollback after duplicate error: %v", rbErr)
			}
			return false, nil
		}
		tx.Rollback()
		return false, err
	}

	var bal model.Balance
	err := tx.
		Where("user_id = ? AND wallet_id = ?", e.UserID, e.WalletID).
		First(&bal).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return false, err
	}

	if err == gorm.ErrRecordNotFound {
		//  create.
		bal = model.Balance{
			UserID:    e.UserID,
			WalletID:  e.WalletID,
			Balance:   e.WalletBalance,
			UpdatedAt: e.UpdatedAt,
		}
		if err := tx.Create(&bal).Error; err != nil {
			tx.Rollback()
			return false, err
		}
		if err := tx.Commit().Error; err != nil {
			return false, err
		}
		return true, nil
	}

	// only update if an event is newer.
	if !e.UpdatedAt.After(bal.UpdatedAt) {
		if err := tx.Commit().Error; err != nil {
			return false, err
		}
		log.Printf("repository: stale event %s for user=%d wallet=%d (event time %v, latest %v) – ignored",
			e.EventID, e.UserID, e.WalletID, e.UpdatedAt, bal.UpdatedAt)
		return false, nil
	}

	bal.Balance = e.WalletBalance
	bal.UpdatedAt = e.UpdatedAt
	if err := tx.Save(&bal).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Commit().Error; err != nil {
		return false, err
	}
	return true, nil
}

func (r *repo) LatestStates() ([]LatestBalance, error) {
	var rows []LatestBalance
	if err := r.db.
		Model(&model.Balance{}).
		Select("user_id, wallet_id, balance, updated_at").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
