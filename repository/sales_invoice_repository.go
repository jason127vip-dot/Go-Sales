package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrDocumentState = errors.New("document is not in the required state")
var ErrAlreadyInvoiced = errors.New("this sales outbound already has an invoice")
var ErrDocumentInUse = errors.New("document has dependent records; remove or reverse them first")
var ErrLegacyPayment = errors.New("this order has confirmed legacy payments; cancel their confirmation and assign them to invoices before confirming invoice payments")
var ErrInvoiceRequired = errors.New("select a confirmed invoice from the same sales order")

func lockSalesOrder(tx *gorm.DB, id uint) (*model.SalesOrder, error) {
	var order model.SalesOrder
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error
	return &order, err
}

// Transaction-scoped numbering also works after deletion and across concurrent requests.
func nextDocumentNumber(tx *gorm.DB, table, column, prefix string) (string, error) {
	if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(83001)).Error; err != nil {
		return "", err
	}
	base := fmt.Sprintf("%s-%d-", prefix, time.Now().Year())
	var next int64
	err := tx.Table(table).Where(column+" LIKE ?", base+"%").Select("COALESCE(MAX(CAST(SPLIT_PART(" + column + ", '-', 3) AS BIGINT)), 0) + 1").Scan(&next).Error
	return fmt.Sprintf("%s%04d", base, next), err
}

func invoiceDetails(tx *gorm.DB) *gorm.DB {
	return tx.Preload("SalesOrder.Customer").Preload("SalesOutbound").Preload("Lines")
}

func invoiceBalance(tx *gorm.DB, invoice *model.SalesInvoice) error {
	if err := tx.Model(&model.Payment{}).Where("sales_invoice_id = ? AND status = ?", invoice.ID, "confirmed").Select("COALESCE(SUM(amount), 0)").Scan(&invoice.PaidAmount).Error; err != nil {
		return err
	}
	invoice.UnpaidAmount = math.Max(0, math.Round((invoice.TotalAmount-invoice.PaidAmount)*100)/100)
	return nil
}

func (r *SalesDocumentRepository) FindInvoices(ctx context.Context) ([]model.SalesInvoice, error) {
	rows := make([]model.SalesInvoice, 0)
	tx := r.db.WithContext(ctx)
	if err := invoiceDetails(tx).Order("id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		if err := invoiceBalance(tx, &rows[i]); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func (r *SalesDocumentRepository) FindInvoice(ctx context.Context, id uint) (*model.SalesInvoice, error) {
	var row model.SalesInvoice
	tx := r.db.WithContext(ctx)
	if err := invoiceDetails(tx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, invoiceBalance(tx, &row)
}

func (r *SalesDocumentRepository) InvoiceOutbounds(ctx context.Context) ([]model.SalesOutbound, error) {
	rows := make([]model.SalesOutbound, 0)
	err := r.db.WithContext(ctx).Preload("SalesOrder.Customer").Preload("Lines.SalesOrderLine").Where("status = ? AND NOT EXISTS (SELECT 1 FROM sales_invoices WHERE sales_outbound_id = sales_outbounds.id)", "confirmed").Order("id desc").Find(&rows).Error
	return rows, err
}

func buildInvoice(outbound model.SalesOutbound, date time.Time, remarks string) model.SalesInvoice {
	order := outbound.SalesOrder
	row := model.SalesInvoice{SalesOutboundID: outbound.ID, SalesOrderID: outbound.SalesOrderID, InvoiceDate: date, CustomerName: order.Customer.Name, CustomerAddress: order.Customer.Address, CustomerPONo: order.CustomerPONo, PaymentTerms: order.Customer.PaymentTerms, Remarks: remarks, Status: "draft", Lines: make([]model.SalesInvoiceLine, 0, len(outbound.Lines))}
	for _, source := range outbound.Lines {
		line := source.SalesOrderLine
		amount := math.Round(source.OutboundQuantity*line.UnitPrice*100) / 100
		row.Lines = append(row.Lines, model.SalesInvoiceLine{ProductCode: line.ProductCode, ProductName: line.ProductName, Specification: line.Specification, Unit: line.Unit, Quantity: source.OutboundQuantity, UnitPrice: line.UnitPrice, Amount: amount})
		row.TotalAmount += amount
	}
	row.TotalAmount = math.Round(row.TotalAmount*100) / 100
	return row
}

func (r *SalesDocumentRepository) CreateInvoice(ctx context.Context, req dto.SaveSalesInvoiceRequest) (*model.SalesInvoice, error) {
	date, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		return nil, err
	}
	var row model.SalesInvoice
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var outbound model.SalesOutbound
		if err := tx.First(&outbound, req.SalesOutboundID).Error; err != nil {
			return err
		}
		order, err := lockSalesOrder(tx, outbound.SalesOrderID)
		if err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("SalesOrder.Customer").Preload("Lines.SalesOrderLine").First(&outbound, req.SalesOutboundID).Error; err != nil {
			return err
		}
		if order.Status != "confirmed" || outbound.Status != "confirmed" || len(outbound.Lines) == 0 {
			return ErrDocumentState
		}
		var count int64
		if err := tx.Model(&model.SalesInvoice{}).Where("sales_outbound_id = ?", outbound.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrAlreadyInvoiced
		}
		row = buildInvoice(outbound, date, req.Remarks)
		row.InvoiceNo, err = nextDocumentNumber(tx, "sales_invoices", "invoice_no", "INV")
		if err != nil {
			return err
		}
		return tx.Create(&row).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindInvoice(ctx, row.ID)
}

func (r *SalesDocumentRepository) changeInvoice(ctx context.Context, id uint, change func(*gorm.DB, *model.SalesInvoice) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.SalesInvoice
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

func (r *SalesDocumentRepository) UpdateInvoice(ctx context.Context, id uint, req dto.SaveSalesInvoiceRequest) (*model.SalesInvoice, error) {
	date, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		return nil, err
	}
	err = r.changeInvoice(ctx, id, func(tx *gorm.DB, row *model.SalesInvoice) error {
		if row.Status != "draft" || row.SalesOutboundID != req.SalesOutboundID {
			return ErrDocumentState
		}
		return tx.Model(row).Updates(map[string]any{"invoice_date": date, "remarks": req.Remarks}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindInvoice(ctx, id)
}

func (r *SalesDocumentRepository) SetInvoiceStatus(ctx context.Context, id uint, from, to string) (*model.SalesInvoice, error) {
	err := r.changeInvoice(ctx, id, func(tx *gorm.DB, row *model.SalesInvoice) error {
		if row.Status != from {
			return ErrDocumentState
		}
		if to == "draft" {
			var count int64
			if err := tx.Model(&model.Payment{}).Where("sales_invoice_id = ?", id).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrDocumentInUse
			}
		} else {
			var outbound model.SalesOutbound
			if err := tx.First(&outbound, row.SalesOutboundID).Error; err != nil {
				return err
			}
			var order model.SalesOrder
			if err := tx.First(&order, row.SalesOrderID).Error; err != nil {
				return err
			}
			if outbound.Status != "confirmed" || order.Status != "confirmed" {
				return ErrDocumentState
			}
		}
		return tx.Model(row).Update("status", to).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindInvoice(ctx, id)
}

func (r *SalesDocumentRepository) DeleteInvoice(ctx context.Context, id uint) error {
	return r.changeInvoice(ctx, id, func(tx *gorm.DB, row *model.SalesInvoice) error {
		if row.Status != "draft" {
			return ErrDocumentState
		}
		if err := tx.Where("sales_invoice_id = ?", id).Delete(&model.SalesInvoiceLine{}).Error; err != nil {
			return err
		}
		return tx.Delete(row).Error
	})
}
