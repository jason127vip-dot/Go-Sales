package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/repository"
	"github.com/jason127vip-dot/Go-Sales/response"
	"github.com/jason127vip-dot/Go-Sales/service"
)

type PriceListHandler struct{ service *service.PriceListService }

func NewPriceListHandler(s *service.PriceListService) *PriceListHandler {
	return &PriceListHandler{service: s}
}
func (h *PriceListHandler) FindAll(c *gin.Context) {
	rows, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, rows)
}
func (h *PriceListHandler) Save(c *gin.Context) {
	var id uint
	if c.Param("id") != "" {
		parsed, ok := parseID(c)
		if !ok {
			return
		}
		id = parsed
	}
	var input dto.SavePriceListRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	if err := h.service.Save(c.Request.Context(), id, input); err != nil {
		if errors.Is(err, service.ErrInvalidPriceList) || errors.Is(err, repository.ErrPriceListOverlap) {
			response.Error(c, 400, err.Error())
			return
		}
		handleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "price saved"})
}
func (h *PriceListHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "price deleted"})
}
func (h *PriceListHandler) Resolve(c *gin.Context) {
	customerID, customerErr := strconv.ParseUint(c.Query("customerId"), 10, 64)
	productID, productErr := strconv.ParseUint(c.Query("productId"), 10, 64)
	if customerErr != nil || productErr != nil {
		response.Error(c, 400, "customerId and productId are required")
		return
	}
	result, err := h.service.Resolve(c.Request.Context(), uint(customerID), uint(productID), c.Query("date"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidPriceList) {
			response.Error(c, 400, "a valid customer, product and date are required")
			return
		}
		handleError(c, err)
		return
	}
	response.Success(c, result)
}
