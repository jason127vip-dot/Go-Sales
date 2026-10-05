package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/repository"
)

var ErrInvalidPriceList = errors.New("start date must be on or before end date and price must be a non-negative amount with no more than two decimal places")

type PriceListService struct {
	repository *repository.PriceListRepository
}

func NewPriceListService(r *repository.PriceListRepository) *PriceListService {
	return &PriceListService{repository: r}
}
func (s *PriceListService) FindAll(ctx context.Context) ([]model.PriceList, error) {
	return s.repository.FindAll(ctx)
}
func (s *PriceListService) Save(ctx context.Context, id uint, input dto.SavePriceListRequest) error {
	start, startErr := time.Parse("2006-01-02", input.StartDate)
	end, endErr := time.Parse("2006-01-02", input.EndDate)
	if startErr != nil || endErr != nil || start.After(end) || !validPrice(input.UnitPrice) {
		return ErrInvalidPriceList
	}
	return s.repository.Save(ctx, &model.PriceList{ID: id, CustomerID: input.CustomerID, ProductID: input.ProductID, StartDate: start, EndDate: end, UnitPrice: input.UnitPrice})
}
func (s *PriceListService) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}
func (s *PriceListService) Resolve(ctx context.Context, customerID, productID uint, date string) (*repository.ResolvedPrice, error) {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil || customerID == 0 || productID == 0 {
		return nil, ErrInvalidPriceList
	}
	return s.repository.Resolve(ctx, customerID, productID, parsed)
}
func validPrice(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && math.Abs(value*100-math.Round(value*100)) < 0.000001
}
