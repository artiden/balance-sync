package db

import (
	"goservice/internal/config"
	"goservice/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// We shouldn't use that in production, but in the MVP it's ok
	if err := db.AutoMigrate(&model.Balance{}, &model.BalanceEvent{}); err != nil {
		return nil, err
	}

	return db, nil
}
