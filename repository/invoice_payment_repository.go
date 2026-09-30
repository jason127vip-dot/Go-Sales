package repository

import (
	"context"
	"math"
	"time"

	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentInvoiceSummary struct {
	ID               uint    `json:"id"`
	SalesOrderID     uint    `json:"salesOrderId"`
	InvoiceNo        string  `json:"invoiceNo"`
	OrderNo          string  `json:"orderNo"`
	CustomerName     string  `json:"customerName"`
	InvoiceAmount    float64 `json:"invoiceAmount"`
	PaidAmount       float64 `json:"paidAmount"`
	UnpaidAmount     float64 `json:"unpaidAmount"`
	LegacyPaidAmount float64 `json:"legacyPaidAmount"`
}

func (r *SalesDocumentRepository) PaymentSummaries(ctx context.Context) ([]PaymentInvoiceSummary, error) {
	invoices, err := r.FindInvoices(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]PaymentInvoiceSummary, 0)
	for _, invoice := range invoices {
		if invoice.Status != "confirmed" || invoice.UnpaidAmount <= 0 {
			continue
		}
		var legacy float64
		if err := r.db.WithContext(ctx).Model(&model.Payment{}).Where("sales_order_id = ? AND sales_invoice_id IS NULL AND status = ?", invoice.SalesOrderID, "confirmed").Select("COALESCE(SUM(amount), 0)").Scan(&legacy).Error; err != nil {
			return nil, err
		}
		rows = append(rows, PaymentInvoiceSummary{invoice.ID, invoice.SalesOrderID, invoice.InvoiceNo, invoice.SalesOrder.OrderNo, invoice.CustomerName, invoice.TotalAmount, invoice.PaidAmount, invoice.UnpaidAmount, legacy})
	}
	return rows, nil
}

func validateInvoicePayment(tx *gorm.DB, invoice *model.SalesInvoice, amount float64, confirming bool) error {
	if invoice.Status != "confirmed" {
		return ErrInvoiceRequired
	}
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 || math.Abs(amount*100-math.Round(amount*100)) > 0.000001 {
		return ErrInsufficientBalance
	}
	if err := invoiceBalance(tx, invoice); err != nil {
		return err
	}
	if amount > invoice.UnpaidAmount {
		return ErrInsufficientBalance
	}
	if confirming {
		var count int64
		if err := tx.Model(&model.Payment{}).Where("sales_order_id = ? AND sales_invoice_id IS NULL AND status = ?", invoice.SalesOrderID, "confirmed").Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrLegacyPayment
		}
	}
	return nil
}

func (r *SalesDocumentRepository) CreatePayment(ctx context.Context, req dto.CreatePaymentRequest) (*model.Payment, error) {
	date, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, err
	}
	var row model.Payment
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var invoice model.SalesInvoice
		if err := tx.First(&invoice, req.SalesInvoiceID).Error; err != nil {
			return err
		}
		order, err := lockSalesOrder(tx, invoice.SalesOrderID)
		if err != nil {
			return err
		}
		if order.Status != "confirmed" {
			return ErrDocumentState
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&invoice, req.SalesInvoiceID).Error; err != nil {
			return err
		}
		if err := validateInvoicePayment(tx, &invoice, req.Amount, false); err != nil {
			return err
		}
		number, err := nextDocumentNumber(tx, "payments", "payment_no", "PAY")
		if err != nil {
			return err
		}
		row = model.Payment{PaymentNo: number, SalesInvoiceID: &invoice.ID, SalesOrderID: invoice.SalesOrderID, PaymentDate: date, Amount: req.Amount, Method: req.Method, ReferenceNo: req.ReferenceNo, Status: "draft"}
		return tx.Create(&row).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindPayment(ctx, row.ID)
}

func (r *SalesDocumentRepository) changePayment(ctx context.Context, id uint, change func(*gorm.DB, *model.Payment) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.Payment
		if err := tx.First(&row, id).Error; err != nil {
			return err
		}
		if _, err := lockSalesOrder(tx, row.SalesOrderID); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
			return err
		}
		return change(tx, &row)
	})
}

func (r *SalesDocumentRepository) UpdatePayment(ctx context.Context, id uint, req dto.CreatePaymentRequest) (*model.Payment, error) {
	date, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, err
	}
	err = r.changePayment(ctx, id, func(tx *gorm.DB, row *model.Payment) error {
		if row.Status != "draft" {
			return ErrDocumentState
		}
		var invoice model.SalesInvoice
		if err := tx.First(&invoice, req.SalesInvoiceID).Error; err != nil {
			return err
		}
		if invoice.SalesOrderID != row.SalesOrderID || (row.SalesInvoiceID != nil && *row.SalesInvoiceID != invoice.ID) {
			return ErrInvoiceRequired
		}
		if err := validateInvoicePayment(tx, &invoice, req.Amount, false); err != nil {
			return err
		}
		return tx.Model(row).Updates(map[string]any{"sales_invoice_id": invoice.ID, "payment_date": date, "amount": req.Amount, "method": req.Method, "reference_no": req.ReferenceNo}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindPayment(ctx, id)
}

func (r *SalesDocumentRepository) SetPaymentStatus(ctx context.Context, id uint, from, to string) (*model.Payment, error) {
	err := r.changePayment(ctx, id, func(tx *gorm.DB, row *model.Payment) error {
		if row.Status != from {
			return ErrDocumentState
		}
		if to == "confirmed" {
			if row.SalesInvoiceID == nil {
				return ErrInvoiceRequired
			}
			var invoice model.SalesInvoice
			if err := tx.First(&invoice, *row.SalesInvoiceID).Error; err != nil {
				return err
			}
			var order model.SalesOrder
			if err := tx.First(&order, row.SalesOrderID).Error; err != nil {
				return err
			}
			if order.Status != "confirmed" {
				return ErrDocumentState
			}
			if err := validateInvoicePayment(tx, &invoice, row.Amount, true); err != nil {
				return err
			}
		}
		return tx.Model(row).Update("status", to).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindPayment(ctx, id)
}

func (r *SalesDocumentRepository) DeletePayment(ctx context.Context, id uint) error {
	return r.changePayment(ctx, id, func(tx *gorm.DB, row *model.Payment) error {
		if row.Status != "draft" {
			return ErrDocumentState
		}
		return tx.Delete(row).Error
	})
}

func (r *SalesDocumentRepository) SetOutboundStatus(ctx context.Context, id uint, from, to string) (*model.SalesOutbound, error) {
	var result *model.SalesOutbound
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		if to == "draft" {
			var count int64
			if err := tx.Model(&model.SalesInvoice{}).Where("sales_outbound_id = ?", id).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrDocumentInUse
			}
		}
		var err error
		result, err = NewSalesDocumentRepository(tx).setOutboundStatus(ctx, id, from, to)
		return err
	})
	return result, err
}
