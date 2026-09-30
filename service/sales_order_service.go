package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/repository"
)

type SalesOrderService interface {
	FindAll(context.Context) ([]model.SalesOrder, error)
	Create(context.Context, dto.CreateSalesOrderRequest) (*model.SalesOrder, error)
	Update(context.Context, uint, dto.UpdateSalesOrderRequest) (*model.SalesOrder, error)
	Confirm(context.Context, uint) (*model.SalesOrder, error)
	CancelConfirmation(context.Context, uint) (*model.SalesOrder, error)
	Delete(context.Context, uint) error
}

type SalesOrderServiceImpl struct {
	repository repository.SalesOrderRepository
}

func NewSalesOrderService(repository repository.SalesOrderRepository) *SalesOrderServiceImpl {
	return &SalesOrderServiceImpl{repository: repository}
}

func (s *SalesOrderServiceImpl) FindAll(ctx context.Context) ([]model.SalesOrder, error) {
	return s.repository.FindAll(ctx)
}
func (s *SalesOrderServiceImpl) Create(ctx context.Context, req dto.CreateSalesOrderRequest) (*model.SalesOrder, error) {
	order, err := orderFromRequest(req)
	if err != nil {
		return nil, err
	}
	return s.repository.Create(ctx, order)
}
func (s *SalesOrderServiceImpl) Update(ctx context.Context, id uint, req dto.UpdateSalesOrderRequest) (*model.SalesOrder, error) {
	order, err := orderFromRequest(req)
	if err != nil {
		return nil, err
	}
	order.ID = id
	return s.repository.UpdateDraft(ctx, order)
}
func (s *SalesOrderServiceImpl) Confirm(ctx context.Context, id uint) (*model.SalesOrder, error) {
	return s.repository.Confirm(ctx, id)
}
func (s *SalesOrderServiceImpl) CancelConfirmation(ctx context.Context, id uint) (*model.SalesOrder, error) {
	return s.repository.CancelConfirmation(ctx, id)
}
func (s *SalesOrderServiceImpl) Delete(ctx context.Context, id uint) error {
	return s.repository.DeleteDraft(ctx, id)
}

func orderFromRequest(req dto.CreateSalesOrderRequest) (*model.SalesOrder, error) {
	orderDate, err := time.Parse("2006-01-02", req.OrderDate)
	if err != nil {
		return nil, fmt.Errorf("invalid order date")
	}
	var expectedOutboundDate *time.Time
	if req.ExpectedOutboundDate != "" {
		parsed, parseErr := time.Parse("2006-01-02", req.ExpectedOutboundDate)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid expected outbound date")
		}
		expectedOutboundDate = &parsed
	}
	lines := make([]model.SalesOrderLine, 0, len(req.Lines))
	for _, line := range req.Lines {
		lines = append(lines, model.SalesOrderLine{ProductID: line.ProductID, Quantity: line.Quantity})
	}
	return &model.SalesOrder{CustomerID: req.CustomerID, OrderDate: orderDate, CustomerPONo: req.CustomerPONo, ExpectedOutboundDate: expectedOutboundDate, Salesperson: req.Salesperson, Remarks: req.Remarks, Status: "draft", Lines: lines}, nil
}
