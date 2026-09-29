package model

import "time"

type SalesOutbound struct {
	ID           uint                `json:"id" gorm:"primaryKey;autoIncrement"`
	OutboundNo   string              `json:"outboundNo" gorm:"size:50;not null;uniqueIndex"`
	SalesOrderID uint                `json:"salesOrderId" gorm:"not null;index"`
	SalesOrder   SalesOrder          `json:"salesOrder" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	OutboundDate time.Time           `json:"outboundDate" gorm:"type:date;not null"`
	Status       string              `json:"status" gorm:"size:20;not null;default:draft;index"`
	Lines        []SalesOutboundLine `json:"lines" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`
}

type SalesOutboundLine struct {
	ID               uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	SalesOutboundID  uint           `json:"salesOutboundId" gorm:"not null;index"`
	SalesOrderLineID uint           `json:"salesOrderLineId" gorm:"not null;index"`
	SalesOrderLine   SalesOrderLine `json:"salesOrderLine" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	OutboundQuantity float64        `json:"outboundQuantity" gorm:"type:numeric(18,4);not null"`
}
