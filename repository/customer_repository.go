package repository

import (
	"context"

	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	FindAll(ctx context.Context) ([]model.Customer, error)
	Create(ctx context.Context, customer *model.Customer) error
	Update(ctx context.Context, customer *model.Customer) (*model.Customer, error)
	Delete(ctx context.Context, id uint) error
}

type CustomerRepositoryImpl struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepositoryImpl {
	return &CustomerRepositoryImpl{db: db}
}

func (r *CustomerRepositoryImpl) FindAll(ctx context.Context) ([]model.Customer, error) {
	var customers []model.Customer
	if err := r.db.WithContext(ctx).Order("id desc").Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *CustomerRepositoryImpl) Create(ctx context.Context, customer *model.Customer) error {
	return r.db.WithContext(ctx).Create(customer).Error
}

func (r *CustomerRepositoryImpl) Update(ctx context.Context, customer *model.Customer) (*model.Customer, error) {
	var existing model.Customer
	if err := r.db.WithContext(ctx).First(&existing, customer.ID).Error; err != nil {
		return nil, err
	}

	existing.CustomerCode = customer.CustomerCode
	existing.Name = customer.Name
	existing.ContactPerson = customer.ContactPerson
	existing.Phone = customer.Phone
	existing.Email = customer.Email
	existing.Address = customer.Address
	existing.PaymentTerms = customer.PaymentTerms
	existing.Status = customer.Status

	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (r *CustomerRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Customer{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
