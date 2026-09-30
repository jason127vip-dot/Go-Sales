package service

import (
	"context"
	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/model"
)

func (s *SalesDocumentService) FindInvoices(c context.Context) ([]model.SalesInvoice, error) {
	return s.repository.FindInvoices(c)
}
func (s *SalesDocumentService) InvoiceOutbounds(c context.Context) ([]model.SalesOutbound, error) {
	return s.repository.InvoiceOutbounds(c)
}
func (s *SalesDocumentService) CreateInvoice(c context.Context, req dto.SaveSalesInvoiceRequest) (*model.SalesInvoice, error) {
	return s.repository.CreateInvoice(c, req)
}
func (s *SalesDocumentService) UpdateInvoice(c context.Context, id uint, req dto.SaveSalesInvoiceRequest) (*model.SalesInvoice, error) {
	return s.repository.UpdateInvoice(c, id, req)
}
func (s *SalesDocumentService) ConfirmInvoice(c context.Context, id uint) (*model.SalesInvoice, error) {
	return s.repository.SetInvoiceStatus(c, id, "draft", "confirmed")
}
func (s *SalesDocumentService) CancelInvoice(c context.Context, id uint) (*model.SalesInvoice, error) {
	return s.repository.SetInvoiceStatus(c, id, "confirmed", "draft")
}
func (s *SalesDocumentService) DeleteInvoice(c context.Context, id uint) error {
	return s.repository.DeleteInvoice(c, id)
}
