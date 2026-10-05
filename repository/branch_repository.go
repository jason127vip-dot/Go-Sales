package repository

import (
	"context"
	"errors"

	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrBranchCodeInUse = errors.New("branch code cannot be changed after sales orders have been created")

type BranchRepository struct{ db *gorm.DB }

func NewBranchRepository(db *gorm.DB) *BranchRepository { return &BranchRepository{db: db} }
func (r *BranchRepository) FindAll(ctx context.Context) ([]model.Branch, error) {
	rows := make([]model.Branch, 0)
	err := r.db.WithContext(ctx).Order("id").Find(&rows).Error
	return rows, err
}
func (r *BranchRepository) Save(ctx context.Context, row *model.Branch) error {
	if row.ID == 0 {
		return r.db.WithContext(ctx).Create(row).Error
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.Branch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&existing, row.ID).Error; err != nil {
			return err
		}
		if existing.Code != row.Code {
			var count int64
			if err := tx.Model(&model.SalesOrder{}).Where("branch_id = ?", row.ID).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrBranchCodeInUse
			}
		}
		var allocated float64
		if err := tx.Model(&model.BranchCustomerCredit{}).Where("branch_id = ?", row.ID).Select("COALESCE(SUM(credit_limit), 0)").Scan(&allocated).Error; err != nil {
			return err
		}
		if roundMoney(allocated) > row.TotalCreditLimit {
			return ErrCreditAllocationExceeded
		}
		return tx.Model(&existing).Updates(map[string]any{"code": row.Code, "name": row.Name, "enable_credit_control": row.EnableCreditControl, "total_credit_limit": row.TotalCreditLimit}).Error
	})
}

// Downstream documents inherit their branch through the sales order.
func branchScope(ctx context.Context, table string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		id := model.BranchID(ctx)
		if id == 0 {
			return db
		}
		if table == "sales_orders" {
			return db.Where("sales_orders.branch_id = ?", id)
		}
		return db.Where(table+".sales_order_id IN (SELECT id FROM sales_orders WHERE branch_id = ?)", id)
	}
}
