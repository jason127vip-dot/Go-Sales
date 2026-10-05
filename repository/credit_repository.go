package repository

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrCreditAllocationExceeded = errors.New("customer credit allocations exceed the branch total credit limit")

type CreditLimitExceededError struct {
	Limit     float64
	Used      float64
	Available float64
	Order     float64
}

func (e *CreditLimitExceededError) Error() string {
	exceeded := math.Max(0, e.Order-e.Available)
	return fmt.Sprintf("credit limit exceeded: limit $%.2f, used $%.2f, available $%.2f, this order $%.2f, exceeded by $%.2f", e.Limit, e.Used, e.Available, e.Order, exceeded)
}

type CustomerCreditRow struct {
	CustomerID      uint    `json:"customerId"`
	CustomerCode    string  `json:"customerCode"`
	CustomerName    string  `json:"customerName"`
	CustomerStatus  string  `json:"customerStatus"`
	CreditLimit     float64 `json:"creditLimit"`
	UsedCredit      float64 `json:"usedCredit"`
	AvailableCredit float64 `json:"availableCredit"`
	Status          string  `json:"status"`
}

type CreditSummary struct {
	BranchID            uint                `json:"branchId"`
	EnableCreditControl bool                `json:"enableCreditControl"`
	TotalCreditLimit    float64             `json:"totalCreditLimit"`
	AllocatedCredit     float64             `json:"allocatedCredit"`
	UnallocatedCredit   float64             `json:"unallocatedCredit"`
	Customers           []CustomerCreditRow `json:"customers"`
}

type CreditRepository struct{ db *gorm.DB }

func NewCreditRepository(db *gorm.DB) *CreditRepository { return &CreditRepository{db: db} }

func (r *CreditRepository) Summary(ctx context.Context) (*CreditSummary, error) {
	branchID := model.BranchID(ctx)
	var branch model.Branch
	if err := r.db.WithContext(ctx).First(&branch, branchID).Error; err != nil {
		return nil, err
	}
	var customers []model.Customer
	if err := r.db.WithContext(ctx).Order("customer_code").Find(&customers).Error; err != nil {
		return nil, err
	}
	var credits []model.BranchCustomerCredit
	if err := r.db.WithContext(ctx).Where("branch_id = ?", branchID).Find(&credits).Error; err != nil {
		return nil, err
	}
	limits := make(map[uint]float64, len(credits))
	for _, credit := range credits {
		limits[credit.CustomerID] = credit.CreditLimit
	}
	result := &CreditSummary{BranchID: branchID, EnableCreditControl: branch.EnableCreditControl, TotalCreditLimit: branch.TotalCreditLimit, Customers: make([]CustomerCreditRow, 0, len(customers))}
	for _, customer := range customers {
		limit := limits[customer.ID]
		used, err := customerUsedCredit(r.db.WithContext(ctx), branchID, customer.ID)
		if err != nil {
			return nil, err
		}
		available := roundMoney(limit - used)
		status := "available"
		if _, configured := limits[customer.ID]; !configured {
			status = "not_configured"
		} else if available < 0 {
			status = "exceeded"
		}
		result.AllocatedCredit += limit
		result.Customers = append(result.Customers, CustomerCreditRow{CustomerID: customer.ID, CustomerCode: customer.CustomerCode, CustomerName: customer.Name, CustomerStatus: customer.Status, CreditLimit: limit, UsedCredit: used, AvailableCredit: available, Status: status})
	}
	result.AllocatedCredit = roundMoney(result.AllocatedCredit)
	result.UnallocatedCredit = roundMoney(branch.TotalCreditLimit - result.AllocatedCredit)
	return result, nil
}

func (r *CreditRepository) SetCustomerLimit(ctx context.Context, customerID uint, limit float64) error {
	branchID := model.BranchID(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var branch model.Branch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&branch, branchID).Error; err != nil {
			return err
		}
		var customer model.Customer
		if err := tx.First(&customer, customerID).Error; err != nil {
			return err
		}
		var allocated float64
		if err := tx.Model(&model.BranchCustomerCredit{}).Where("branch_id = ? AND customer_id <> ?", branchID, customerID).Select("COALESCE(SUM(credit_limit), 0)").Scan(&allocated).Error; err != nil {
			return err
		}
		if roundMoney(allocated+limit) > branch.TotalCreditLimit {
			return ErrCreditAllocationExceeded
		}
		row := model.BranchCustomerCredit{BranchID: branchID, CustomerID: customerID, CreditLimit: limit}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "branch_id"}, {Name: "customer_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"credit_limit", "updated_at"}),
		}).Create(&row).Error
	})
}

type creditPosition struct {
	Enabled   bool
	Limit     float64
	Used      float64
	Available float64
}

func orderCreditPosition(tx *gorm.DB, ctx context.Context, order *model.SalesOrder, lock bool) (*creditPosition, error) {
	branchID := model.BranchID(ctx)
	var branch model.Branch
	branchQuery := tx
	if lock {
		branchQuery = branchQuery.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := branchQuery.First(&branch, branchID).Error; err != nil {
		return nil, err
	}
	position := &creditPosition{Enabled: branch.EnableCreditControl}
	if !position.Enabled {
		return position, nil
	}
	query := tx.Where("branch_id = ? AND customer_id = ?", branchID, order.CustomerID)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var credit model.BranchCustomerCredit
	err := query.First(&credit).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		position.Limit = credit.CreditLimit
	}
	used, err := customerUsedCredit(tx, branchID, order.CustomerID)
	if err != nil {
		return nil, err
	}
	position.Used = used
	position.Available = roundMoney(position.Limit - position.Used)
	return position, nil
}

func creditWarningOrError(tx *gorm.DB, ctx context.Context, order *model.SalesOrder, strict bool) (string, error) {
	position, err := orderCreditPosition(tx, ctx, order, strict)
	if err != nil {
		return "", err
	}
	return evaluateCredit(*position, roundMoney(order.TotalAmount), strict)
}

func evaluateCredit(position creditPosition, orderAmount float64, strict bool) (string, error) {
	if !position.Enabled || orderAmount <= position.Available {
		return "", nil
	}
	creditError := &CreditLimitExceededError{Limit: position.Limit, Used: position.Used, Available: position.Available, Order: orderAmount}
	if strict {
		return "", creditError
	}
	return "Draft saved. Warning: " + creditError.Error() + ". The order cannot be confirmed unless credit becomes available.", nil
}

func customerUsedCredit(tx *gorm.DB, branchID, customerID uint) (float64, error) {
	var orders float64
	if err := tx.Model(&model.SalesOrder{}).Where("branch_id = ? AND customer_id = ? AND status = ?", branchID, customerID, "confirmed").Select("COALESCE(SUM(total_amount), 0)").Scan(&orders).Error; err != nil {
		return 0, err
	}
	var payments float64
	if err := tx.Model(&model.Payment{}).
		Joins("JOIN sales_orders ON sales_orders.id = payments.sales_order_id").
		Where("sales_orders.branch_id = ? AND sales_orders.customer_id = ? AND payments.status = ?", branchID, customerID, "confirmed").
		Select("COALESCE(SUM(payments.amount), 0)").Scan(&payments).Error; err != nil {
		return 0, err
	}
	return roundMoney(math.Max(0, orders-payments)), nil
}

func roundMoney(value float64) float64 { return math.Round(value*100) / 100 }
