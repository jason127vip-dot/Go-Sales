package model

import "time"

type PriceList struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	BranchID   uint      `json:"branchId" gorm:"not null;uniqueIndex:idx_price_list_range;index"`
	CustomerID uint      `json:"customerId" gorm:"not null;uniqueIndex:idx_price_list_range;index"`
	Customer   Customer  `json:"customer" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ProductID  uint      `json:"productId" gorm:"not null;uniqueIndex:idx_price_list_range;index"`
	Product    Product   `json:"product" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	StartDate  time.Time `json:"startDate" gorm:"type:date;not null;uniqueIndex:idx_price_list_range"`
	EndDate    time.Time `json:"endDate" gorm:"type:date;not null;uniqueIndex:idx_price_list_range"`
	UnitPrice  float64   `json:"unitPrice" gorm:"type:numeric(18,2);not null"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
