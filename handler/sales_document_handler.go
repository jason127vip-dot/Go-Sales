package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/repository"
	"github.com/jason127vip-dot/Go-Sales/response"
	"github.com/jason127vip-dot/Go-Sales/service"
)

type SalesDocumentHandler struct{ service *service.SalesDocumentService }

func NewSalesDocumentHandler(service *service.SalesDocumentService) *SalesDocumentHandler {
	return &SalesDocumentHandler{service}
}
func (h *SalesDocumentHandler) FindOutbounds(c *gin.Context) {
	rows, e := h.service.FindOutbounds(c.Request.Context())
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, rows)
}
func (h *SalesDocumentHandler) AvailableOrders(c *gin.Context) {
	rows, e := h.service.AvailableOrders(c.Request.Context())
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, rows)
}
func (h *SalesDocumentHandler) OutboundLines(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	rows, e := h.service.OutboundLines(c.Request.Context(), id)
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, rows)
}
func (h *SalesDocumentHandler) CreateOutbound(c *gin.Context) {
	var req dto.CreateSalesOutboundRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	row, e := h.service.CreateOutbound(c.Request.Context(), req)
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) UpdateOutbound(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CreateSalesOutboundRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	row, e := h.service.UpdateOutbound(c.Request.Context(), id, req)
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) ConfirmOutbound(c *gin.Context) { h.outboundStatus(c, true) }
func (h *SalesDocumentHandler) CancelOutbound(c *gin.Context)  { h.outboundStatus(c, false) }
func (h *SalesDocumentHandler) outboundStatus(c *gin.Context, confirm bool) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var row any
	var e error
	if confirm {
		row, e = h.service.ConfirmOutbound(c.Request.Context(), id)
	} else {
		row, e = h.service.CancelOutbound(c.Request.Context(), id)
	}
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) DeleteOutbound(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if e := h.service.DeleteOutbound(c.Request.Context(), id); e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, gin.H{"message": "sales outbound deleted"})
}
func (h *SalesDocumentHandler) FindPayments(c *gin.Context) {
	rows, e := h.service.FindPayments(c.Request.Context())
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, rows)
}
func (h *SalesDocumentHandler) PaymentSummaries(c *gin.Context) {
	rows, e := h.service.PaymentSummaries(c.Request.Context())
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, rows)
}
func (h *SalesDocumentHandler) CreatePayment(c *gin.Context) {
	var req dto.CreatePaymentRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	row, e := h.service.CreatePayment(c.Request.Context(), req)
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) UpdatePayment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CreatePaymentRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	row, e := h.service.UpdatePayment(c.Request.Context(), id, req)
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) ConfirmPayment(c *gin.Context) { h.paymentStatus(c, true) }
func (h *SalesDocumentHandler) CancelPayment(c *gin.Context)  { h.paymentStatus(c, false) }
func (h *SalesDocumentHandler) paymentStatus(c *gin.Context, confirm bool) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var row any
	var e error
	if confirm {
		row, e = h.service.ConfirmPayment(c.Request.Context(), id)
	} else {
		row, e = h.service.CancelPayment(c.Request.Context(), id)
	}
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) DeletePayment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if e := h.service.DeletePayment(c.Request.Context(), id); e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, gin.H{"message": "payment deleted"})
}
func (h *SalesDocumentHandler) Execution(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	row, e := h.service.Execution(c.Request.Context(), id)
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) PaymentReport(c *gin.Context) {
	rows, e := h.service.PaymentReport(c.Request.Context())
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, rows)
}
func (h *SalesDocumentHandler) ARAgingReport(c *gin.Context) {
	rows, e := h.service.ARAgingReport(c.Request.Context())
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, rows)
}
func (h *SalesDocumentHandler) Dashboard(c *gin.Context) {
	row, e := h.service.Dashboard(c.Request.Context())
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func handleDocumentError(c *gin.Context, e error) {
	if errors.Is(e, repository.ErrDocumentState) || errors.Is(e, repository.ErrAlreadyInvoiced) || errors.Is(e, repository.ErrDocumentInUse) || errors.Is(e, repository.ErrLegacyPayment) || errors.Is(e, repository.ErrInvoiceRequired) || errors.Is(e, repository.ErrInsufficientBalance) || errors.Is(e, repository.ErrInsufficientQuantity) || errors.Is(e, repository.ErrOnlyDraftOrders) || errors.Is(e, repository.ErrOnlyConfirmedOrders) {
		response.Error(c, 400, e.Error())
		return
	}
	handleError(c, e)
}
