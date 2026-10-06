package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInsufficientQuantity = errors.New("outbound quantity exceeds the remaining order quantity")
var ErrInsufficientBalance = errors.New("payment amount exceeds the unpaid amount")

type SalesDocumentRepository struct{ db *gorm.DB }

type OutboundLineAvailability struct {
	ID                uint    `json:"id"`
	SalesOrderLineID  uint    `json:"salesOrderLineId"`
	ProductCode       string  `json:"productCode"`
	ProductName       string  `json:"productName"`
	Unit              string  `json:"unit"`
	OrderedQuantity   float64 `json:"orderedQuantity"`
	RemainingQuantity float64 `json:"remainingQuantity"`
	OutboundQuantity  float64 `json:"outboundQuantity"`
}

type SalesOrderExecution struct {
	Outbounds []model.SalesOutbound `json:"outbounds"`
	Invoices  []model.SalesInvoice  `json:"invoices"`
	Payments  []model.Payment       `json:"payments"`
}

type SalesOrderPaymentReportRow struct {
	ID              uint    `json:"id"`
	OrderNo         string  `json:"orderNo"`
	CustomerName    string  `json:"customerName"`
	OrderDate       string  `json:"orderDate"`
	OrderAmount     float64 `json:"orderAmount"`
	PaidAmount      float64 `json:"paidAmount"`
	UnpaidAmount    float64 `json:"unpaidAmount"`
	LastPaymentDate *string `json:"lastPaymentDate"`
	PaymentStatus   string  `json:"paymentStatus"`
}
type ARAgingReportRow struct {
	ID                uint    `json:"id"`
	InvoiceNo         string  `json:"invoiceNo"`
	OrderNo           string  `json:"orderNo"`
	CustomerName      string  `json:"customerName"`
	InvoiceDate       string  `json:"invoiceDate"`
	InvoiceAmount     float64 `json:"invoiceAmount"`
	PaidAmount        float64 `json:"paidAmount"`
	OutstandingAmount float64 `json:"outstandingAmount"`
	AgingDays         int     `json:"agingDays"`
	AgingBucket       string  `json:"agingBucket"`
}
type DashboardSalesOrder struct {
	ID           uint    `json:"id"`
	OrderNo      string  `json:"orderNo"`
	CustomerName string  `json:"customerName"`
	OrderDate    string  `json:"orderDate"`
	Amount       float64 `json:"amount"`
	Status       string  `json:"status"`
}
type OutstandingCustomer struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}
type DailyOrderVolume struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
type DashboardStats struct {
	TotalSales           float64               `json:"totalSales"`
	ReceivedAmount       float64               `json:"receivedAmount"`
	UnpaidAmount         float64               `json:"unpaidAmount"`
	OutboundQuantity     float64               `json:"outboundQuantity"`
	RecentSalesOrders    []DashboardSalesOrder `json:"recentSalesOrders"`
	OutstandingCustomers []OutstandingCustomer `json:"outstandingCustomers"`
	DailyOrderVolume     []DailyOrderVolume    `json:"dailyOrderVolume"`
}

func NewSalesDocumentRepository(db *gorm.DB) *SalesDocumentRepository {
	return &SalesDocumentRepository{db: db}
}

func (r *SalesDocumentRepository) CreateOutbound(ctx context.Context, req dto.CreateSalesOutboundRequest) (*model.SalesOutbound, error) {
	date, err := time.Parse("2006-01-02", req.OutboundDate)
	if err != nil {
		return nil, err
	}
	var result model.SalesOutbound
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := lockSalesOrder(tx, req.SalesOrderID)
		if err != nil {
			return err
		}
		if order.Status != "confirmed" {
			return ErrOnlyConfirmedOrders
		}
		lines := make([]model.SalesOutboundLine, 0, len(req.Lines))
		for _, reqLine := range req.Lines {
			var orderLine model.SalesOrderLine
			if err := tx.Where("id = ? AND sales_order_id = ?", reqLine.SalesOrderLineID, order.ID).First(&orderLine).Error; err != nil {
				return err
			}
			var sent float64
			tx.Model(&model.SalesOutboundLine{}).Joins("JOIN sales_outbounds ON sales_outbounds.id = sales_outbound_lines.sales_outbound_id").Where("sales_outbound_lines.sales_order_line_id = ? AND sales_outbounds.status = ?", orderLine.ID, "confirmed").Select("COALESCE(SUM(sales_outbound_lines.outbound_quantity), 0)").Scan(&sent)
			if reqLine.OutboundQuantity > orderLine.Quantity-sent {
				return ErrInsufficientQuantity
			}
			lines = append(lines, model.SalesOutboundLine{SalesOrderLineID: orderLine.ID, OutboundQuantity: reqLine.OutboundQuantity})
		}
		number, err := nextDocumentNumber(tx, "sales_outbounds", "outbound_no", "OUT")
		if err != nil {
			return err
		}
		result = model.SalesOutbound{OutboundNo: number, SalesOrderID: order.ID, OutboundDate: date, Status: "draft", Lines: lines}
		return tx.Create(&result).Error
	})
	return &result, err
}

func (r *SalesDocumentRepository) FindOutbounds(ctx context.Context) ([]model.SalesOutbound, error) {
	var rows []model.SalesOutbound
	err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_outbounds")).Preload("SalesOrder.Customer").Preload("Lines.SalesOrderLine").Order("id desc").Find(&rows).Error
	return rows, err
}

func (r *SalesDocumentRepository) AvailableOrders(ctx context.Context) ([]model.SalesOrder, error) {
	var orders []model.SalesOrder
	if err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_orders")).Preload("Customer").Preload("Lines").Where("status = ?", "confirmed").Order("id desc").Find(&orders).Error; err != nil {
		return nil, err
	}
	result := make([]model.SalesOrder, 0)
	for _, order := range orders {
		lines, err := r.OutboundLines(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		if len(lines) > 0 {
			result = append(result, order)
		}
	}
	return result, nil
}

func (r *SalesDocumentRepository) OutboundLines(ctx context.Context, orderID uint) ([]OutboundLineAvailability, error) {
	var order model.SalesOrder
	if err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_orders")).Preload("Lines").Where("id = ? AND status = ?", orderID, "confirmed").First(&order).Error; err != nil {
		return nil, err
	}
	result := make([]OutboundLineAvailability, 0)
	for _, line := range order.Lines {
		var sent float64
		err := r.db.WithContext(ctx).Table("sales_outbound_lines").Joins("JOIN sales_outbounds ON sales_outbounds.id = sales_outbound_lines.sales_outbound_id").Where("sales_outbound_lines.sales_order_line_id = ? AND sales_outbounds.status = ?", line.ID, "confirmed").Select("COALESCE(SUM(sales_outbound_lines.outbound_quantity), 0)").Scan(&sent).Error
		if err != nil {
			return nil, err
		}
		remaining := line.Quantity - sent
		if remaining > 0 {
			result = append(result, OutboundLineAvailability{line.ID, line.ID, line.ProductCode, line.ProductName, line.Unit, line.Quantity, remaining, remaining})
		}
	}
	return result, nil
}

func (r *SalesDocumentRepository) setOutboundStatus(ctx context.Context, id uint, from, to string) (*model.SalesOutbound, error) {
	var row model.SalesOutbound
	if err := r.db.WithContext(ctx).Preload("Lines").First(&row, id).Error; err != nil {
		return nil, err
	}
	if row.Status != from {
		return nil, ErrOnlyDraftOrders
	}
	if to == "confirmed" {
		available, err := r.OutboundLines(ctx, row.SalesOrderID)
		if err != nil {
			return nil, err
		}
		for _, saved := range row.Lines {
			valid := false
			for _, item := range available {
				if item.SalesOrderLineID == saved.SalesOrderLineID && saved.OutboundQuantity <= item.RemainingQuantity {
					valid = true
					break
				}
			}
			if !valid {
				return nil, ErrInsufficientQuantity
			}
		}
	}
	row.Status = to
	if err := r.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, err
	}
	return r.FindOutbound(ctx, id)
}

func (r *SalesDocumentRepository) FindOutbound(ctx context.Context, id uint) (*model.SalesOutbound, error) {
	var row model.SalesOutbound
	err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_outbounds")).Preload("SalesOrder.Customer").Preload("Lines.SalesOrderLine").First(&row, id).Error
	return &row, err
}

func (r *SalesDocumentRepository) UpdateOutbound(ctx context.Context, id uint, req dto.CreateSalesOutboundRequest) (*model.SalesOutbound, error) {
	date, err := time.Parse("2006-01-02", req.OutboundDate)
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.SalesOutbound
		if err := tx.First(&row, id).Error; err != nil {
			return err
		}
		if _, err := lockSalesOrder(tx, row.SalesOrderID); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
			return err
		}
		if row.Status != "draft" || row.SalesOrderID != req.SalesOrderID {
			return ErrOnlyDraftOrders
		}
		available, err := NewSalesDocumentRepository(tx).OutboundLines(ctx, row.SalesOrderID)
		if err != nil {
			return err
		}
		lines := make([]model.SalesOutboundLine, 0, len(req.Lines))
		for _, input := range req.Lines {
			found := false
			for _, item := range available {
				if item.SalesOrderLineID == input.SalesOrderLineID {
					found = true
					if input.OutboundQuantity > item.RemainingQuantity {
						return ErrInsufficientQuantity
					}
					lines = append(lines, model.SalesOutboundLine{SalesOutboundID: id, SalesOrderLineID: input.SalesOrderLineID, OutboundQuantity: input.OutboundQuantity})
					break
				}
			}
			if !found {
				return ErrInsufficientQuantity
			}
		}
		if err := tx.Where("sales_outbound_id = ?", id).Delete(&model.SalesOutboundLine{}).Error; err != nil {
			return err
		}
		row.OutboundDate = date
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		return tx.Create(&lines).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindOutbound(ctx, id)
}

func (r *SalesDocumentRepository) DeleteOutbound(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.SalesOutbound
		if err := tx.First(&row, id).Error; err != nil {
			return err
		}
		if _, err := lockSalesOrder(tx, row.SalesOrderID); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
			return err
		}
		if row.Status != "draft" {
			return ErrOnlyDraftOrders
		}
		if err := tx.Where("sales_outbound_id = ?", id).Delete(&model.SalesOutboundLine{}).Error; err != nil {
			return err
		}
		return tx.Delete(&row).Error
	})
}

func (r *SalesDocumentRepository) FindPayments(ctx context.Context) ([]model.Payment, error) {
	var rows []model.Payment
	err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "payments")).Preload("SalesOrder.Customer").Preload("SalesInvoice").Order("id desc").Find(&rows).Error
	return rows, err
}

func (r *SalesDocumentRepository) FindPayment(ctx context.Context, id uint) (*model.Payment, error) {
	var row model.Payment
	err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "payments")).Preload("SalesOrder.Customer").Preload("SalesInvoice").First(&row, id).Error
	return &row, err
}

func (r *SalesDocumentRepository) Execution(ctx context.Context, orderID uint) (*SalesOrderExecution, error) {
	var outbounds []model.SalesOutbound
	if err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_outbounds")).Preload("SalesOrder.Customer").Preload("Lines.SalesOrderLine").Where("sales_order_id = ?", orderID).Order("id desc").Find(&outbounds).Error; err != nil {
		return nil, err
	}
	var payments []model.Payment
	if err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "payments")).Preload("SalesOrder.Customer").Preload("SalesInvoice").Where("sales_order_id = ?", orderID).Order("id desc").Find(&payments).Error; err != nil {
		return nil, err
	}
	invoices := make([]model.SalesInvoice, 0)
	if err := invoiceDetails(r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_invoices"))).Where("sales_order_id = ?", orderID).Order("id desc").Find(&invoices).Error; err != nil {
		return nil, err
	}
	for i := range invoices {
		if err := invoiceBalance(r.db.WithContext(ctx), &invoices[i]); err != nil {
			return nil, err
		}
	}
	return &SalesOrderExecution{Outbounds: outbounds, Invoices: invoices, Payments: payments}, nil
}

func (r *SalesDocumentRepository) PaymentReport(ctx context.Context) ([]SalesOrderPaymentReportRow, error) {
	var orders []model.SalesOrder
	if err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_orders")).Preload("Customer").Where("status = ?", "confirmed").Order("order_date desc, id desc").Find(&orders).Error; err != nil {
		return nil, err
	}
	rows := make([]SalesOrderPaymentReportRow, 0, len(orders))
	for _, order := range orders {
		var payments []model.Payment
		if err := r.db.WithContext(ctx).Where("sales_order_id = ? AND status = ?", order.ID, "confirmed").Order("payment_date desc, id desc").Find(&payments).Error; err != nil {
			return nil, err
		}
		paid := 0.0
		for _, payment := range payments {
			paid += payment.Amount
		}
		unpaid := order.TotalAmount - paid
		status := "partially_paid"
		if paid == 0 {
			status = "unpaid"
		} else if unpaid <= 0 {
			unpaid = 0
			status = "paid"
		}
		var last *string
		if len(payments) > 0 {
			value := payments[0].PaymentDate.Format("2006-01-02")
			last = &value
		}
		rows = append(rows, SalesOrderPaymentReportRow{order.ID, order.OrderNo, order.Customer.Name, order.OrderDate.Format("2006-01-02"), order.TotalAmount, paid, unpaid, last, status})
	}
	return rows, nil
}

func (r *SalesDocumentRepository) ARAgingReport(ctx context.Context) ([]ARAgingReportRow, error) {
	invoices, err := r.FindInvoices(ctx)
	if err != nil {
		return nil, err
	}
	today, err := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	rows := make([]ARAgingReportRow, 0)
	for _, invoice := range invoices {
		if invoice.Status != "confirmed" || invoice.UnpaidAmount <= 0 {
			continue
		}
		days := invoiceAgingDays(invoice.InvoiceDate, today)
		rows = append(rows, ARAgingReportRow{
			ID: invoice.ID, InvoiceNo: invoice.InvoiceNo, OrderNo: invoice.SalesOrder.OrderNo,
			CustomerName: invoice.CustomerName, InvoiceDate: invoice.InvoiceDate.Format("2006-01-02"),
			InvoiceAmount: invoice.TotalAmount, PaidAmount: invoice.PaidAmount, OutstandingAmount: invoice.UnpaidAmount,
			AgingDays: days, AgingBucket: agingBucket(days),
		})
	}
	return rows, nil
}

func invoiceAgingDays(invoiceDate, analysisDate time.Time) int {
	days := int(analysisDate.Sub(invoiceDate).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

func agingBucket(days int) string {
	switch {
	case days <= 30:
		return "0-30 Days"
	case days <= 60:
		return "31-60 Days"
	case days <= 90:
		return "61-90 Days"
	case days <= 120:
		return "91-120 Days"
	default:
		return "120+ Days"
	}
}

func (r *SalesDocumentRepository) Dashboard(ctx context.Context) (*DashboardStats, error) {
	report, err := r.PaymentReport(ctx)
	if err != nil {
		return nil, err
	}
	stats := &DashboardStats{
		RecentSalesOrders:    make([]DashboardSalesOrder, 0),
		OutstandingCustomers: make([]OutstandingCustomer, 0),
		DailyOrderVolume:     make([]DailyOrderVolume, 0, 14),
	}
	balances := map[string]float64{}
	for _, row := range report {
		stats.TotalSales += row.OrderAmount
		stats.ReceivedAmount += row.PaidAmount
		stats.UnpaidAmount += row.UnpaidAmount
		if row.UnpaidAmount > 0 {
			balances[row.CustomerName] += row.UnpaidAmount
		}
	}
	if err := r.db.WithContext(ctx).Table("sales_outbound_lines").Joins("JOIN sales_outbounds ON sales_outbounds.id = sales_outbound_lines.sales_outbound_id").Scopes(branchScope(ctx, "sales_outbounds")).Where("sales_outbounds.status = ?", "confirmed").Select("COALESCE(SUM(sales_outbound_lines.outbound_quantity), 0)").Scan(&stats.OutboundQuantity).Error; err != nil {
		return nil, err
	}
	var recent []model.SalesOrder
	if err := r.db.WithContext(ctx).Scopes(branchScope(ctx, "sales_orders")).Preload("Customer").Order("created_at desc").Limit(5).Find(&recent).Error; err != nil {
		return nil, err
	}
	for _, order := range recent {
		stats.RecentSalesOrders = append(stats.RecentSalesOrders, DashboardSalesOrder{
			ID:           order.ID,
			OrderNo:      order.OrderNo,
			CustomerName: order.Customer.Name,
			OrderDate:    order.OrderDate.Format("2006-01-02"),
			Amount:       order.TotalAmount,
			Status:       order.Status,
		})
	}
	for name, amount := range balances {
		stats.OutstandingCustomers = append(stats.OutstandingCustomers, OutstandingCustomer{name, amount})
	}
	sort.Slice(stats.OutstandingCustomers, func(i, j int) bool {
		return stats.OutstandingCustomers[i].Amount > stats.OutstandingCustomers[j].Amount
	})

	today := time.Now()
	startDate := today.AddDate(0, 0, -13)
	type dailyOrderCount struct {
		Date  string
		Count int64
	}
	var dailyCounts []dailyOrderCount
	if err := r.db.WithContext(ctx).Model(&model.SalesOrder{}).Scopes(branchScope(ctx, "sales_orders")).
		Select("TO_CHAR(order_date, 'YYYY-MM-DD') AS date, COUNT(*) AS count").
		Where("order_date BETWEEN ? AND ?", startDate.Format("2006-01-02"), today.Format("2006-01-02")).
		Group("order_date").
		Order("order_date").
		Scan(&dailyCounts).Error; err != nil {
		return nil, err
	}
	countsByDate := make(map[string]int64, len(dailyCounts))
	for _, item := range dailyCounts {
		countsByDate[item.Date] = item.Count
	}
	for day := 0; day < 14; day++ {
		date := startDate.AddDate(0, 0, day).Format("2006-01-02")
		stats.DailyOrderVolume = append(stats.DailyOrderVolume, DailyOrderVolume{Date: date, Count: countsByDate[date]})
	}
	return stats, nil
}
