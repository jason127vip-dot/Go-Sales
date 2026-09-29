package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/response"
	"github.com/jason127vip-dot/Go-Sales/service"
)

type ProductHandler struct {
	service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) FindAll(c *gin.Context) {
	products, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, products)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	product := productFromRequest(req)
	if err := h.service.Create(c.Request.Context(), &product); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	product := productFromRequest(req)
	product.ID = id
	updated, err := h.service.Update(c.Request.Context(), &product)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, updated)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "product deleted"})
}

func productFromRequest(req dto.CreateProductRequest) model.Product {
	product := model.Product{
		ProductCode: req.ProductCode, Name: req.Name, Specification: req.Specification,
		Unit: req.Unit, UnitPrice: req.UnitPrice, Status: req.Status,
	}
	if barcode := strings.TrimSpace(req.Barcode); barcode != "" {
		product.Barcode = &barcode
	}
	return product
}
