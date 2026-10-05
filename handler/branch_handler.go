package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/repository"
	"github.com/jason127vip-dot/Go-Sales/response"
	"github.com/jason127vip-dot/Go-Sales/service"
	"gorm.io/gorm"
)

type BranchHandler struct{ service *service.BranchService }

func NewBranchHandler(s *service.BranchService) *BranchHandler { return &BranchHandler{service: s} }
func (h *BranchHandler) FindAll(c *gin.Context) {
	rows, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, rows)
}
func (h *BranchHandler) Save(c *gin.Context) {
	var input struct {
		Code                string  `json:"code"`
		Name                string  `json:"name"`
		EnableCreditControl bool    `json:"enableCreditControl"`
		TotalCreditLimit    float64 `json:"totalCreditLimit"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, "invalid branch details")
		return
	}
	row := model.Branch{Code: input.Code, Name: input.Name, EnableCreditControl: input.EnableCreditControl, TotalCreditLimit: input.TotalCreditLimit}
	if c.Param("id") != "" {
		id, ok := parseID(c)
		if !ok {
			return
		}
		row.ID = id
	}
	if err := h.service.Save(c.Request.Context(), &row); err != nil {
		if errors.Is(err, service.ErrInvalidBranch) || errors.Is(err, repository.ErrBranchCodeInUse) || errors.Is(err, repository.ErrCreditAllocationExceeded) {
			response.Error(c, 400, err.Error())
			return
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			response.Error(c, 400, "branch code already exists")
			return
		}
		handleError(c, err)
		return
	}
	response.Success(c, row)
}
