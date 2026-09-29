package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/response"
	"github.com/jason127vip-dot/Go-Sales/service"
)

type CustomerHandler struct {
	service service.CustomerService
}

func NewCustomerHandler(service service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) FindAll(c *gin.Context) {
	customers, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, customers)
}

func (h *CustomerHandler) Create(c *gin.Context) {
	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	customer := model.Customer{
		CustomerCode: req.CustomerCode, Name: req.Name, ContactPerson: req.ContactPerson,
		Phone: req.Phone, Email: req.Email, Address: req.Address,
		PaymentTerms: req.PaymentTerms, Status: req.Status,
	}
	if err := h.service.Create(c.Request.Context(), &customer); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, customer)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	customer := model.Customer{
		ID: id, CustomerCode: req.CustomerCode, Name: req.Name, ContactPerson: req.ContactPerson,
		Phone: req.Phone, Email: req.Email, Address: req.Address,
		PaymentTerms: req.PaymentTerms, Status: req.Status,
	}
	updated, err := h.service.Update(c.Request.Context(), &customer)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, updated)
}

func (h *CustomerHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "customer deleted"})
}
