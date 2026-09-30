package repository

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/jason127vip-dot/Go-Sales/model"
)

func TestBuildInvoiceUsesOutboundQuantityAndOrderSnapshot(t *testing.T) {
	source := model.SalesOutbound{ID: 7, SalesOrderID: 3, SalesOrder: model.SalesOrder{
		Customer: model.Customer{Name: "Original customer", Address: "Original address", PaymentTerms: "30 days"}, CustomerPONo: "PO-123",
	}, Lines: []model.SalesOutboundLine{
		{OutboundQuantity: 1.25, SalesOrderLine: model.SalesOrderLine{Quantity: 10, ProductCode: "P1", ProductName: "Original product", Specification: "Small", Unit: "ea", UnitPrice: 1.99}},
		{OutboundQuantity: 3, SalesOrderLine: model.SalesOrderLine{Quantity: 20, ProductCode: "P2", UnitPrice: 0.1}},
	}}
	date := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	invoice := buildInvoice(source, date, "Handle carefully")
	if invoice.TotalAmount != 2.79 || invoice.Lines[0].Amount != 2.49 || invoice.Lines[1].Amount != 0.3 {
		t.Fatalf("unexpected rounded amounts: %+v", invoice)
	}
	if invoice.SalesOrderID != 3 || invoice.SalesOutboundID != 7 || invoice.Lines[0].Quantity != 1.25 || invoice.Status != "draft" || invoice.InvoiceDate != date {
		t.Fatal("incorrect source or initial state")
	}
	source.SalesOrder.Customer.Name = "Changed customer"
	source.Lines[0].SalesOrderLine.ProductName = "Changed product"
	if invoice.CustomerName != "Original customer" || invoice.CustomerAddress != "Original address" || invoice.PaymentTerms != "30 days" || invoice.CustomerPONo != "PO-123" || invoice.Lines[0].ProductName != "Original product" || invoice.Lines[0].Specification != "Small" || invoice.Remarks != "Handle carefully" {
		t.Fatal("invoice did not retain its snapshot")
	}
}

func TestInvoicePaymentRejectsInvalidAmountBeforeDatabaseAccess(t *testing.T) {
	for _, amount := range []float64{0, -1, 0.001, 1.999, math.NaN(), math.Inf(1)} {
		if err := validateInvoicePayment(nil, &model.SalesInvoice{Status: "confirmed"}, amount, false); !errors.Is(err, ErrInsufficientBalance) {
			t.Fatalf("amount %v: got %v", amount, err)
		}
	}
	if err := validateInvoicePayment(nil, &model.SalesInvoice{Status: "draft"}, 10, false); !errors.Is(err, ErrInvoiceRequired) {
		t.Fatalf("draft invoice accepted: %v", err)
	}
}
