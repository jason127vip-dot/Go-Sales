package service

import (
	"context"

	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/repository"
)

type ProductService interface {
	FindAll(ctx context.Context) ([]model.Product, error)
	Create(ctx context.Context, product *model.Product) error
	Update(ctx context.Context, product *model.Product) (*model.Product, error)
	Delete(ctx context.Context, id uint) error
}

type ProductServiceImpl struct {
	repository repository.ProductRepository
}

func NewProductService(repository repository.ProductRepository) *ProductServiceImpl {
	return &ProductServiceImpl{repository: repository}
}

func (s *ProductServiceImpl) FindAll(ctx context.Context) ([]model.Product, error) {
	return s.repository.FindAll(ctx)
}

func (s *ProductServiceImpl) Create(ctx context.Context, product *model.Product) error {
	return s.repository.Create(ctx, product)
}

func (s *ProductServiceImpl) Update(ctx context.Context, product *model.Product) (*model.Product, error) {
	return s.repository.Update(ctx, product)
}

func (s *ProductServiceImpl) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}
