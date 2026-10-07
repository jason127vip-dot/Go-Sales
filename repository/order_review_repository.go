package repository

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
)

type OrderReviewRepository struct{ db *gorm.DB }

type ReviewInvoice struct {
	InvoiceNo         string  `json:"invoiceNo"`
	InvoiceDate       string  `json:"invoiceDate"`
	OutstandingAmount float64 `json:"outstandingAmount"`
	AgingDays         int     `json:"agingDays"`
	AgingBucket       string  `json:"agingBucket"`
}

type OrderReviewContext struct {
	OrderID              uint            `json:"orderId"`
	OrderNo              string          `json:"orderNo"`
	CustomerID           uint            `json:"customerId"`
	CustomerCode         string          `json:"customerCode"`
	CustomerName         string          `json:"customerName"`
	OrderAmount          float64         `json:"orderAmount"`
	CreditControlEnabled bool            `json:"creditControlEnabled"`
	CreditConfigured     bool            `json:"creditConfigured"`
	CreditLimit          float64         `json:"creditLimit"`
	UsedCredit           float64         `json:"usedCredit"`
	AvailableCredit      float64         `json:"availableCredit"`
	Invoices             []ReviewInvoice `json:"invoices"`
}

func NewOrderReviewRepository(db *gorm.DB) *OrderReviewRepository {
	return &OrderReviewRepository{db: db}
}

func (r *OrderReviewRepository) Context(ctx context.Context, id uint) (*OrderReviewContext, error) {
	var order model.SalesOrder
	if err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_orders")).Preload("Customer").First(&order, id).Error; err != nil {
		return nil, err
	}
	if order.Status != "draft" {
		return nil, ErrOnlyDraftOrders
	}

	var branch model.Branch
	if err := r.db.WithContext(ctx).First(&branch, model.BranchID(ctx)).Error; err != nil {
		return nil, err
	}
	var credit model.BranchCustomerCredit
	creditErr := r.db.WithContext(ctx).Where("branch_id = ? AND customer_id = ?", branch.ID, order.CustomerID).First(&credit).Error
	if creditErr != nil && !errors.Is(creditErr, gorm.ErrRecordNotFound) {
		return nil, creditErr
	}
	used, err := customerUsedCredit(r.db.WithContext(ctx), branch.ID, order.CustomerID)
	if err != nil {
		return nil, err
	}

	result := &OrderReviewContext{
		OrderID: order.ID, OrderNo: order.OrderNo, CustomerID: order.CustomerID,
		CustomerCode: order.Customer.CustomerCode, CustomerName: order.Customer.Name, OrderAmount: order.TotalAmount,
		CreditControlEnabled: branch.EnableCreditControl, CreditConfigured: creditErr == nil,
		CreditLimit: credit.CreditLimit, UsedCredit: used, AvailableCredit: roundMoney(credit.CreditLimit - used),
		Invoices: make([]ReviewInvoice, 0),
	}

	var invoices []model.SalesInvoice
	query := invoiceDetails(r.db.WithContext(ctx)).Scopes(branchScope(ctx, "sales_invoices")).
		Where("sales_invoices.status = ? AND sales_invoices.sales_order_id IN (SELECT id FROM sales_orders WHERE customer_id = ?)", "confirmed", order.CustomerID).
		Order("sales_invoices.invoice_date, sales_invoices.id")
	if err := query.Find(&invoices).Error; err != nil {
		return nil, err
	}
	today, err := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	for i := range invoices {
		if err := invoiceBalance(r.db.WithContext(ctx), &invoices[i]); err != nil {
			return nil, err
		}
		if invoices[i].UnpaidAmount <= 0 {
			continue
		}
		days := invoiceAgingDays(invoices[i].InvoiceDate, today)
		result.Invoices = append(result.Invoices, ReviewInvoice{
			InvoiceNo: invoices[i].InvoiceNo, InvoiceDate: invoices[i].InvoiceDate.Format("2006-01-02"),
			OutstandingAmount: math.Round(invoices[i].UnpaidAmount*100) / 100,
			AgingDays:         days, AgingBucket: agingBucket(days),
		})
	}
	return result, nil
}
