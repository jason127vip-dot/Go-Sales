package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/repository"
	"github.com/jason127vip-dot/Go-Sales/response"
	"github.com/jason127vip-dot/Go-Sales/service"
)

type SalesOrderHandler struct{ service service.SalesOrderService }

func NewSalesOrderHandler(service service.SalesOrderService) *SalesOrderHandler {
	return &SalesOrderHandler{service: service}
}

func (h *SalesOrderHandler) FindAll(c *gin.Context) {
	orders, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, orders)
}

func (h *SalesOrderHandler) Create(c *gin.Context) {
	var req dto.CreateSalesOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	order, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSalesOrderError(c, err)
		return
	}
	response.Success(c, order)
}

func (h *SalesOrderHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateSalesOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	order, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleSalesOrderError(c, err)
		return
	}
	response.Success(c, order)
}

func (h *SalesOrderHandler) Confirm(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	order, err := h.service.Confirm(c.Request.Context(), id)
	if err != nil {
		handleSalesOrderError(c, err)
		return
	}
	response.Success(c, order)
}

func (h *SalesOrderHandler) CancelConfirmation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	order, err := h.service.CancelConfirmation(c.Request.Context(), id)
	if err != nil {
		handleSalesOrderError(c, err)
		return
	}
	response.Success(c, order)
}

func (h *SalesOrderHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleSalesOrderError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "sales order deleted"})
}

func handleSalesOrderError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrOnlyDraftOrders) || errors.Is(err, repository.ErrOnlyConfirmedOrders) {
		response.Error(c, 400, err.Error())
		return
	}
	handleError(c, err)
}
