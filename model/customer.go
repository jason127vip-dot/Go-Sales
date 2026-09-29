package model

import "time"

type Customer struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CustomerCode  string    `json:"customerCode" gorm:"size:50;not null;uniqueIndex"`
	Name          string    `json:"name" gorm:"size:200;not null"`
	ContactPerson string    `json:"contactPerson" gorm:"size:100"`
	Phone         string    `json:"phone" gorm:"size:50"`
	Email         string    `json:"email" gorm:"size:150"`
	Address       string    `json:"address" gorm:"size:500"`
	PaymentTerms  string    `json:"paymentTerms" gorm:"size:100"`
	Status        string    `json:"status" gorm:"size:20;not null;default:active"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
