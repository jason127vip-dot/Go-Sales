package repository

import (
	"context"

	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
)

type ProductRepository interface {
	FindAll(ctx context.Context) ([]model.Product, error)
	Create(ctx context.Context, product *model.Product) error
	Update(ctx context.Context, product *model.Product) (*model.Product, error)
	Delete(ctx context.Context, id uint) error
}

type ProductRepositoryImpl struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepositoryImpl {
	return &ProductRepositoryImpl{db: db}
}

func (r *ProductRepositoryImpl) FindAll(ctx context.Context) ([]model.Product, error) {
	var products []model.Product
	if err := r.db.WithContext(ctx).Order("id desc").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepositoryImpl) Create(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *ProductRepositoryImpl) Update(ctx context.Context, product *model.Product) (*model.Product, error) {
	var existing model.Product
	if err := r.db.WithContext(ctx).First(&existing, product.ID).Error; err != nil {
		return nil, err
	}

	existing.ProductCode = product.ProductCode
	existing.Barcode = product.Barcode
	existing.Name = product.Name
	existing.Specification = product.Specification
	existing.Unit = product.Unit
	existing.UnitPrice = product.UnitPrice
	existing.Status = product.Status

	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (r *ProductRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
