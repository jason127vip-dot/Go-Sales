package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/repository"
	"github.com/jason127vip-dot/Go-Sales/response"
	"github.com/jason127vip-dot/Go-Sales/service"
)

type CreditHandler struct{ service *service.CreditService }

func NewCreditHandler(service *service.CreditService) *CreditHandler {
	return &CreditHandler{service: service}
}

func (h *CreditHandler) Summary(c *gin.Context) {
	result, err := h.service.Summary(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *CreditHandler) SetCustomerLimit(c *gin.Context) {
	customerID, ok := parseID(c)
	if !ok {
		return
	}
	var input struct {
		CreditLimit float64 `json:"creditLimit"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, "invalid credit limit")
		return
	}
	if err := h.service.SetCustomerLimit(c.Request.Context(), customerID, input.CreditLimit); err != nil {
		if errors.Is(err, service.ErrInvalidCreditLimit) || errors.Is(err, repository.ErrCreditAllocationExceeded) {
			response.Error(c, 400, err.Error())
			return
		}
		handleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "customer credit limit saved"})
}
