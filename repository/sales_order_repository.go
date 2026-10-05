package repository

import (
	"context"
	"errors"

	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
)

var ErrOnlyDraftOrders = errors.New("only draft sales orders can be changed")
var ErrOnlyConfirmedOrders = errors.New("only confirmed sales orders can be cancelled")

type SalesOrderRepository interface {
	FindAll(ctx context.Context) ([]model.SalesOrder, error)
	Create(ctx context.Context, order *model.SalesOrder) (*model.SalesOrder, error)
	UpdateDraft(ctx context.Context, order *model.SalesOrder) (*model.SalesOrder, error)
	Confirm(ctx context.Context, id uint) (*model.SalesOrder, error)
	CancelConfirmation(ctx context.Context, id uint) (*model.SalesOrder, error)
	DeleteDraft(ctx context.Context, id uint) error
}

type SalesOrderRepositoryImpl struct{ db *gorm.DB }

func NewSalesOrderRepository(db *gorm.DB) *SalesOrderRepositoryImpl {
	return &SalesOrderRepositoryImpl{db: db}
}

func (r *SalesOrderRepositoryImpl) withDetails(ctx context.Context, query *gorm.DB) *gorm.DB {
	return query.WithContext(ctx).Scopes(branchScope(ctx, "sales_orders")).Preload("Branch").Preload("Customer").Preload("Lines")
}

func (r *SalesOrderRepositoryImpl) FindAll(ctx context.Context) ([]model.SalesOrder, error) {
	var orders []model.SalesOrder
	if err := r.withDetails(ctx, r.db).Order("id desc").Find(&orders).Error; err != nil {
		return nil, err
	}
	for index := range orders {
		if err := r.populateExecution(ctx, &orders[index]); err != nil {
			return nil, err
		}
	}
	return orders, nil
}

func (r *SalesOrderRepositoryImpl) Create(ctx context.Context, order *model.SalesOrder) (*model.SalesOrder, error) {
	order.BranchID = model.BranchID(ctx)
	var creditWarning string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := prepareOrderLines(tx, order); err != nil {
			return err
		}
		var err error
		creditWarning, err = creditWarningOrError(tx, ctx, order, false)
		if err != nil {
			return err
		}
		order.OrderNo, err = nextDocumentNumber(tx, "sales_orders", "order_no", "SO")
		if err != nil {
			return err
		}
		return tx.Create(order).Error
	})
	if err != nil {
		return nil, err
	}
	result, err := r.findByID(ctx, order.ID)
	if result != nil {
		result.CreditWarning = creditWarning
	}
	return result, err
}

func (r *SalesOrderRepositoryImpl) UpdateDraft(ctx context.Context, order *model.SalesOrder) (*model.SalesOrder, error) {
	var creditWarning string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.SalesOrder
		if err := tx.Scopes(branchScope(ctx, "sales_orders")).First(&existing, order.ID).Error; err != nil {
			return err
		}
		if existing.Status != "draft" {
			return ErrOnlyDraftOrders
		}
		if err := prepareOrderLines(tx, order); err != nil {
			return err
		}
		var err error
		creditWarning, err = creditWarningOrError(tx, ctx, order, false)
		if err != nil {
			return err
		}
		if err := tx.Where("sales_order_id = ?", order.ID).Delete(&model.SalesOrderLine{}).Error; err != nil {
			return err
		}
		existing.CustomerID, existing.OrderDate, existing.TotalAmount = order.CustomerID, order.OrderDate, order.TotalAmount
		existing.CustomerPONo, existing.ExpectedOutboundDate = order.CustomerPONo, order.ExpectedOutboundDate
		existing.Salesperson, existing.Remarks = order.Salesperson, order.Remarks
		if err := tx.Save(&existing).Error; err != nil {
			return err
		}
		for index := range order.Lines {
			order.Lines[index].SalesOrderID = existing.ID
		}
		return tx.Create(&order.Lines).Error
	})
	if err != nil {
		return nil, err
	}
	result, err := r.findByID(ctx, order.ID)
	if result != nil {
		result.CreditWarning = creditWarning
	}
	return result, err
}

func (r *SalesOrderRepositoryImpl) Confirm(ctx context.Context, id uint) (*model.SalesOrder, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := lockSalesOrder(tx, id)
		if err != nil {
			return err
		}
		if order.Status != "draft" {
			return ErrOnlyDraftOrders
		}
		if _, err := creditWarningOrError(tx, ctx, order, true); err != nil {
			return err
		}
		return tx.Model(order).Update("status", "confirmed").Error
	})
	if err != nil {
		return nil, err
	}
	return r.findByID(ctx, id)
}

func (r *SalesOrderRepositoryImpl) CancelConfirmation(ctx context.Context, id uint) (*model.SalesOrder, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := lockSalesOrder(tx, id)
		if err != nil {
			return err
		}
		if order.Status != "confirmed" {
			return ErrOnlyConfirmedOrders
		}
		for _, document := range []any{&model.SalesOutbound{}, &model.SalesInvoice{}, &model.Payment{}} {
			var count int64
			if err := tx.Model(document).Where("sales_order_id = ?", id).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrDocumentInUse
			}
		}
		return tx.Model(order).Update("status", "draft").Error
	})
	if err != nil {
		return nil, err
	}
	return r.findByID(ctx, id)
}

func (r *SalesOrderRepositoryImpl) DeleteDraft(ctx context.Context, id uint) error {
	var order model.SalesOrder
	if err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_orders")).First(&order, id).Error; err != nil {
		return err
	}
	if order.Status != "draft" {
		return ErrOnlyDraftOrders
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("sales_order_id = ?", id).Delete(&model.SalesOrderLine{}).Error; err != nil {
			return err
		}
		return tx.Delete(&order).Error
	})
}

func (r *SalesOrderRepositoryImpl) findByID(ctx context.Context, id uint) (*model.SalesOrder, error) {
	var order model.SalesOrder
	if err := r.withDetails(ctx, r.db).First(&order, id).Error; err != nil {
		return nil, err
	}
	if err := r.populateExecution(ctx, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *SalesOrderRepositoryImpl) populateExecution(ctx context.Context, order *model.SalesOrder) error {
	for _, line := range order.Lines {
		order.TotalQuantity += line.Quantity
	}
	var outboundQuantity float64
	if err := r.db.WithContext(ctx).Table("sales_outbound_lines").Joins("JOIN sales_outbounds ON sales_outbounds.id = sales_outbound_lines.sales_outbound_id").Where("sales_outbounds.sales_order_id = ? AND sales_outbounds.status = ?", order.ID, "confirmed").Select("COALESCE(SUM(sales_outbound_lines.outbound_quantity), 0)").Scan(&outboundQuantity).Error; err != nil {
		return err
	}
	order.OutboundStatus = "not_outbound"
	if outboundQuantity > 0 {
		order.OutboundStatus = "partially_outbound"
	}
	if order.TotalQuantity > 0 && outboundQuantity >= order.TotalQuantity {
		order.OutboundStatus = "fully_outbound"
	}
	if err := r.db.WithContext(ctx).Model(&model.Payment{}).Where("sales_order_id = ? AND status = ?", order.ID, "confirmed").Select("COALESCE(SUM(amount), 0)").Scan(&order.PaidAmount).Error; err != nil {
		return err
	}
	order.UnpaidAmount = order.TotalAmount - order.PaidAmount
	if order.UnpaidAmount < 0 {
		order.UnpaidAmount = 0
	}
	order.PaymentStatus = "unpaid"
	if order.PaidAmount > 0 {
		order.PaymentStatus = "partially_paid"
	}
	if order.UnpaidAmount == 0 {
		order.PaymentStatus = "paid"
	}
	return nil
}

func prepareOrderLines(tx *gorm.DB, order *model.SalesOrder) error {
	var customer model.Customer
	if err := tx.Where("id = ? AND status = ?", order.CustomerID, "active").First(&customer).Error; err != nil {
		return err
	}

	order.TotalAmount = 0
	for index := range order.Lines {
		line := &order.Lines[index]
		var product model.Product
		if err := tx.Where("id = ? AND status = ?", line.ProductID, "active").First(&product).Error; err != nil {
			return err
		}
		line.ProductCode = product.ProductCode
		line.ProductName = product.Name
		line.Specification = product.Specification
		line.Unit = product.Unit
		if !line.PriceProvided {
			line.UnitPrice = product.UnitPrice
			var price model.PriceList
			err := tx.Where("branch_id = ? AND customer_id = ? AND product_id = ? AND start_date <= ? AND end_date >= ?", order.BranchID, order.CustomerID, line.ProductID, order.OrderDate, order.OrderDate).First(&price).Error
			if err == nil {
				line.UnitPrice = price.UnitPrice
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		line.Amount = roundMoney(line.UnitPrice * line.Quantity)
		order.TotalAmount += line.Amount
	}
	return nil
}
