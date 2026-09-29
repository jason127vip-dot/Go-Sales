package service

import (
	"context"

	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/repository"
)

type CustomerService interface {
	FindAll(ctx context.Context) ([]model.Customer, error)
	Create(ctx context.Context, customer *model.Customer) error
	Update(ctx context.Context, customer *model.Customer) (*model.Customer, error)
	Delete(ctx context.Context, id uint) error
}

type CustomerServiceImpl struct {
	repository repository.CustomerRepository
}

func NewCustomerService(repository repository.CustomerRepository) *CustomerServiceImpl {
	return &CustomerServiceImpl{repository: repository}
}

func (s *CustomerServiceImpl) FindAll(ctx context.Context) ([]model.Customer, error) {
	return s.repository.FindAll(ctx)
}

func (s *CustomerServiceImpl) Create(ctx context.Context, customer *model.Customer) error {
	return s.repository.Create(ctx, customer)
}

func (s *CustomerServiceImpl) Update(ctx context.Context, customer *model.Customer) (*model.Customer, error) {
	return s.repository.Update(ctx, customer)
}

func (s *CustomerServiceImpl) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}
