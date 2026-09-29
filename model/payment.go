package model

import "time"

type Payment struct {
	ID           uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	PaymentNo    string     `json:"paymentNo" gorm:"size:50;not null;uniqueIndex"`
	SalesOrderID uint       `json:"salesOrderId" gorm:"not null;index"`
	SalesOrder   SalesOrder `json:"salesOrder" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	PaymentDate  time.Time  `json:"paymentDate" gorm:"type:date;not null"`
	Amount       float64    `json:"amount" gorm:"type:numeric(18,2);not null"`
	Method       string     `json:"method" gorm:"size:50;not null"`
	ReferenceNo  string     `json:"referenceNo" gorm:"size:100"`
	Status       string     `json:"status" gorm:"size:20;not null;default:draft;index"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
