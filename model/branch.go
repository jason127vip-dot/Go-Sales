package model

import (
	"context"
	"time"
)

type Branch struct {
	ID                  uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	Code                string  `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Name                string  `json:"name" gorm:"size:200;not null"`
	EnableCreditControl bool    `json:"enableCreditControl" gorm:"not null;default:false"`
	TotalCreditLimit    float64 `json:"totalCreditLimit" gorm:"type:numeric(18,2);not null;default:0"`
}

type BranchCustomerCredit struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	BranchID    uint      `json:"branchId" gorm:"not null;uniqueIndex:idx_branch_customer_credit"`
	Branch      Branch    `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CustomerID  uint      `json:"customerId" gorm:"not null;uniqueIndex:idx_branch_customer_credit"`
	Customer    Customer  `json:"customer" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreditLimit float64   `json:"creditLimit" gorm:"type:numeric(18,2);not null;default:0"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type branchContextKey struct{}

func WithBranch(ctx context.Context, id uint) context.Context {
	return context.WithValue(ctx, branchContextKey{}, id)
}

func BranchID(ctx context.Context) uint {
	id, _ := ctx.Value(branchContextKey{}).(uint)
	return id
}
