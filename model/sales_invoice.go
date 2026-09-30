package model

import "time"

type SalesInvoice struct {
	ID              uint               `json:"id" gorm:"primaryKey;autoIncrement"`
	InvoiceNo       string             `json:"invoiceNo" gorm:"size:50;not null;uniqueIndex"`
	SalesOutboundID uint               `json:"salesOutboundId" gorm:"not null;uniqueIndex"`
	SalesOutbound   SalesOutbound      `json:"salesOutbound" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	SalesOrderID    uint               `json:"salesOrderId" gorm:"not null;index"`
	SalesOrder      SalesOrder         `json:"salesOrder" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	InvoiceDate     time.Time          `json:"invoiceDate" gorm:"type:date;not null"`
	CustomerName    string             `json:"customerName" gorm:"size:200;not null"`
	CustomerAddress string             `json:"customerAddress" gorm:"size:500"`
	CustomerPONo    string             `json:"customerPoNo" gorm:"size:100"`
	PaymentTerms    string             `json:"paymentTerms" gorm:"size:100"`
	Remarks         string             `json:"remarks" gorm:"size:1000"`
	Status          string             `json:"status" gorm:"size:20;not null;default:draft;index"`
	TotalAmount     float64            `json:"totalAmount" gorm:"type:numeric(18,2);not null"`
	Lines           []SalesInvoiceLine `json:"lines" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
	PaidAmount      float64            `json:"paidAmount" gorm:"-"`
	UnpaidAmount    float64            `json:"unpaidAmount" gorm:"-"`
}

type SalesInvoiceLine struct {
	ID             uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	SalesInvoiceID uint    `json:"salesInvoiceId" gorm:"not null;index"`
	ProductCode    string  `json:"productCode" gorm:"size:50;not null"`
	ProductName    string  `json:"productName" gorm:"size:200;not null"`
	Specification  string  `json:"specification" gorm:"size:500"`
	Unit           string  `json:"unit" gorm:"size:30;not null"`
	Quantity       float64 `json:"quantity" gorm:"type:numeric(18,4);not null"`
	UnitPrice      float64 `json:"unitPrice" gorm:"type:numeric(18,2);not null"`
	Amount         float64 `json:"amount" gorm:"type:numeric(18,2);not null"`
}
