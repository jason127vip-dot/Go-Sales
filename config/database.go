package config

import (
	"fmt"

	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Customer{},
		&model.Product{},
		&model.SalesOrder{},
		&model.SalesOrderLine{},
		&model.SalesOutbound{},
		&model.SalesOutboundLine{},
		&model.Payment{},
	)
}
