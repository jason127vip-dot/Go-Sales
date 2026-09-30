package service

import (
	"context"

	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/repository"
)

type SalesDocumentService struct {
	repository *repository.SalesDocumentRepository
}

func NewSalesDocumentService(repository *repository.SalesDocumentRepository) *SalesDocumentService {
	return &SalesDocumentService{repository}
}
func (s *SalesDocumentService) FindOutbounds(c context.Context) ([]model.SalesOutbound, error) {
	return s.repository.FindOutbounds(c)
}
func (s *SalesDocumentService) AvailableOrders(c context.Context) ([]model.SalesOrder, error) {
	return s.repository.AvailableOrders(c)
}
func (s *SalesDocumentService) OutboundLines(c context.Context, id uint) ([]repository.OutboundLineAvailability, error) {
	return s.repository.OutboundLines(c, id)
}
func (s *SalesDocumentService) CreateOutbound(c context.Context, r dto.CreateSalesOutboundRequest) (*model.SalesOutbound, error) {
	return s.repository.CreateOutbound(c, r)
}
func (s *SalesDocumentService) UpdateOutbound(c context.Context, id uint, r dto.CreateSalesOutboundRequest) (*model.SalesOutbound, error) {
	return s.repository.UpdateOutbound(c, id, r)
}
func (s *SalesDocumentService) ConfirmOutbound(c context.Context, id uint) (*model.SalesOutbound, error) {
	return s.repository.SetOutboundStatus(c, id, "draft", "confirmed")
}
func (s *SalesDocumentService) CancelOutbound(c context.Context, id uint) (*model.SalesOutbound, error) {
	return s.repository.SetOutboundStatus(c, id, "confirmed", "draft")
}
func (s *SalesDocumentService) DeleteOutbound(c context.Context, id uint) error {
	return s.repository.DeleteOutbound(c, id)
}
func (s *SalesDocumentService) FindPayments(c context.Context) ([]model.Payment, error) {
	return s.repository.FindPayments(c)
}
func (s *SalesDocumentService) PaymentSummaries(c context.Context) ([]repository.PaymentInvoiceSummary, error) {
	return s.repository.PaymentSummaries(c)
}
func (s *SalesDocumentService) CreatePayment(c context.Context, r dto.CreatePaymentRequest) (*model.Payment, error) {
	return s.repository.CreatePayment(c, r)
}
func (s *SalesDocumentService) UpdatePayment(c context.Context, id uint, r dto.CreatePaymentRequest) (*model.Payment, error) {
	return s.repository.UpdatePayment(c, id, r)
}
func (s *SalesDocumentService) ConfirmPayment(c context.Context, id uint) (*model.Payment, error) {
	return s.repository.SetPaymentStatus(c, id, "draft", "confirmed")
}
func (s *SalesDocumentService) CancelPayment(c context.Context, id uint) (*model.Payment, error) {
	return s.repository.SetPaymentStatus(c, id, "confirmed", "draft")
}
func (s *SalesDocumentService) DeletePayment(c context.Context, id uint) error {
	return s.repository.DeletePayment(c, id)
}
func (s *SalesDocumentService) Execution(c context.Context, id uint) (*repository.SalesOrderExecution, error) {
	return s.repository.Execution(c, id)
}
func (s *SalesDocumentService) PaymentReport(c context.Context) ([]repository.SalesOrderPaymentReportRow, error) {
	return s.repository.PaymentReport(c)
}
func (s *SalesDocumentService) Dashboard(c context.Context) (*repository.DashboardStats, error) {
	return s.repository.Dashboard(c)
}
