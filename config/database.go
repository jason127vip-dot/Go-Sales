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

	dialector := postgres.New(postgres.Config{
		DSN:                  databaseURL,
		PreferSimpleProtocol: true,
	})
	return gorm.Open(dialector, &gorm.Config{TranslateError: true})
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.Branch{}); err != nil {
		return err
	}
	if err := db.Where("id = ?", 1).FirstOrCreate(&model.Branch{Code: "MAIN", Name: "Main Branch"}).Error; err != nil {
		return err
	}
	return db.AutoMigrate(
		&model.Customer{},
		&model.Product{},
		&model.PriceList{},
		&model.BranchCustomerCredit{},
		&model.SalesOrder{},
		&model.SalesOrderLine{},
		&model.SalesOutbound{},
		&model.SalesOutboundLine{},
		&model.SalesInvoice{},
		&model.SalesInvoiceLine{},
		&model.Payment{},
	)
}
