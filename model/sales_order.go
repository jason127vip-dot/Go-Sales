package model

import "time"

type SalesOrder struct {
	ID                   uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderNo              string           `json:"orderNo" gorm:"size:50;not null;uniqueIndex"`
	CustomerID           uint             `json:"customerId" gorm:"not null;index"`
	Customer             Customer         `json:"customer" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	OrderDate            time.Time        `json:"orderDate" gorm:"type:date;not null"`
	CustomerPONo         string           `json:"customerPoNo" gorm:"size:100"`
	ExpectedOutboundDate *time.Time       `json:"expectedOutboundDate" gorm:"type:date"`
	Salesperson          string           `json:"salesperson" gorm:"size:100"`
	Remarks              string           `json:"remarks" gorm:"size:1000"`
	Status               string           `json:"status" gorm:"size:20;not null;default:draft;index"`
	TotalAmount          float64          `json:"totalAmount" gorm:"type:numeric(18,2);not null;default:0"`
	Lines                []SalesOrderLine `json:"lines" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
	TotalQuantity        float64          `json:"totalQuantity" gorm:"-"`
	OutboundStatus       string           `json:"outboundStatus" gorm:"-"`
	PaymentStatus        string           `json:"paymentStatus" gorm:"-"`
	PaidAmount           float64          `json:"paidAmount" gorm:"-"`
	UnpaidAmount         float64          `json:"unpaidAmount" gorm:"-"`
}

type SalesOrderLine struct {
	ID            uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	SalesOrderID  uint    `json:"salesOrderId" gorm:"not null;index"`
	ProductID     uint    `json:"productId" gorm:"not null;index"`
	Product       Product `json:"product" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ProductCode   string  `json:"productCode" gorm:"size:50;not null"`
	ProductName   string  `json:"productName" gorm:"size:200;not null"`
	Specification string  `json:"specification" gorm:"size:500"`
	Unit          string  `json:"unit" gorm:"size:30;not null"`
	UnitPrice     float64 `json:"unitPrice" gorm:"type:numeric(18,2);not null"`
	Quantity      float64 `json:"quantity" gorm:"type:numeric(18,4);not null"`
	Amount        float64 `json:"amount" gorm:"type:numeric(18,2);not null"`
}
