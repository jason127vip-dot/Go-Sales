package service

import (
	"context"
	"errors"
	"math"

	"github.com/jason127vip-dot/Go-Sales/repository"
)

var ErrInvalidCreditLimit = errors.New("credit limit must be a non-negative amount with no more than two decimal places")

type CreditService struct{ repository *repository.CreditRepository }

func NewCreditService(repository *repository.CreditRepository) *CreditService {
	return &CreditService{repository: repository}
}

func (s *CreditService) Summary(ctx context.Context) (*repository.CreditSummary, error) {
	return s.repository.Summary(ctx)
}

func (s *CreditService) SetCustomerLimit(ctx context.Context, customerID uint, limit float64) error {
	if !validCreditAmount(limit) {
		return ErrInvalidCreditLimit
	}
	return s.repository.SetCustomerLimit(ctx, customerID, limit)
}

func validCreditAmount(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && math.Abs(value*100-math.Round(value*100)) < 0.000001
}
