package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrPriceListOverlap = errors.New("a price already exists for this customer and product during the selected date range")

type ResolvedPrice struct {
	UnitPrice float64 `json:"unitPrice"`
	Source    string  `json:"source"`
}

type PriceListRepository struct{ db *gorm.DB }

func NewPriceListRepository(db *gorm.DB) *PriceListRepository { return &PriceListRepository{db: db} }

func (r *PriceListRepository) FindAll(ctx context.Context) ([]model.PriceList, error) {
	rows := make([]model.PriceList, 0)
	err := r.db.WithContext(ctx).Where("price_lists.branch_id = ?", model.BranchID(ctx)).
		Preload("Customer").Preload("Product").Order("start_date desc, id desc").Find(&rows).Error
	return rows, err
}

func (r *PriceListRepository) Save(ctx context.Context, row *model.PriceList) error {
	row.BranchID = model.BranchID(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var branch model.Branch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&branch, row.BranchID).Error; err != nil {
			return err
		}
		if err := tx.First(&model.Customer{}, row.CustomerID).Error; err != nil {
			return err
		}
		if err := tx.First(&model.Product{}, row.ProductID).Error; err != nil {
			return err
		}
		var count int64
		query := tx.Model(&model.PriceList{}).Where(
			"branch_id = ? AND customer_id = ? AND product_id = ? AND start_date <= ? AND end_date >= ?",
			row.BranchID, row.CustomerID, row.ProductID, row.EndDate, row.StartDate,
		)
		if row.ID != 0 {
			query = query.Where("id <> ?", row.ID)
			var existing model.PriceList
			if err := tx.Where("branch_id = ?", row.BranchID).First(&existing, row.ID).Error; err != nil {
				return err
			}
		}
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrPriceListOverlap
		}
		if row.ID == 0 {
			return tx.Create(row).Error
		}
		return tx.Model(&model.PriceList{}).Where("id = ? AND branch_id = ?", row.ID, row.BranchID).
			Updates(map[string]any{"customer_id": row.CustomerID, "product_id": row.ProductID, "start_date": row.StartDate, "end_date": row.EndDate, "unit_price": row.UnitPrice}).Error
	})
}

func (r *PriceListRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ? AND branch_id = ?", id, model.BranchID(ctx)).Delete(&model.PriceList{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *PriceListRepository) Resolve(ctx context.Context, customerID, productID uint, date time.Time) (*ResolvedPrice, error) {
	var row model.PriceList
	err := r.db.WithContext(ctx).Where(
		"branch_id = ? AND customer_id = ? AND product_id = ? AND start_date <= ? AND end_date >= ?",
		model.BranchID(ctx), customerID, productID, date, date,
	).First(&row).Error
	if err == nil {
		return &ResolvedPrice{UnitPrice: row.UnitPrice, Source: "price_list"}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	var product model.Product
	if err := r.db.WithContext(ctx).First(&product, productID).Error; err != nil {
		return nil, err
	}
	return &ResolvedPrice{UnitPrice: product.UnitPrice, Source: "default"}, nil
}
