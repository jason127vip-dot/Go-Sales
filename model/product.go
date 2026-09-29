package model

import "time"

type Product struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductCode   string    `json:"productCode" gorm:"size:50;not null;uniqueIndex"`
	Barcode       *string   `json:"barcode" gorm:"size:100;uniqueIndex"`
	Name          string    `json:"name" gorm:"size:200;not null"`
	Specification string    `json:"specification" gorm:"size:500"`
	Unit          string    `json:"unit" gorm:"size:30;not null"`
	UnitPrice     float64   `json:"unitPrice" gorm:"type:numeric(18,2);not null;default:0"`
	Status        string    `json:"status" gorm:"size:20;not null;default:active"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
