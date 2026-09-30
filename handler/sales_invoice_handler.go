package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/response"
)

func (h *SalesDocumentHandler) FindInvoices(c *gin.Context) {
	rows, e := h.service.FindInvoices(c.Request.Context())
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, rows)
}
func (h *SalesDocumentHandler) InvoiceOutbounds(c *gin.Context) {
	rows, e := h.service.InvoiceOutbounds(c.Request.Context())
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, rows)
}
func (h *SalesDocumentHandler) CreateInvoice(c *gin.Context) {
	var req dto.SaveSalesInvoiceRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	row, e := h.service.CreateInvoice(c.Request.Context(), req)
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) UpdateInvoice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.SaveSalesInvoiceRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	row, e := h.service.UpdateInvoice(c.Request.Context(), id, req)
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) ConfirmInvoice(c *gin.Context) { h.invoiceStatus(c, true) }
func (h *SalesDocumentHandler) CancelInvoice(c *gin.Context)  { h.invoiceStatus(c, false) }
func (h *SalesDocumentHandler) invoiceStatus(c *gin.Context, confirm bool) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var row any
	var e error
	if confirm {
		row, e = h.service.ConfirmInvoice(c.Request.Context(), id)
	} else {
		row, e = h.service.CancelInvoice(c.Request.Context(), id)
	}
	if e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, row)
}
func (h *SalesDocumentHandler) DeleteInvoice(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if e := h.service.DeleteInvoice(c.Request.Context(), id); e != nil {
		handleDocumentError(c, e)
		return
	}
	response.Success(c, gin.H{"message": "sales invoice deleted"})
}
