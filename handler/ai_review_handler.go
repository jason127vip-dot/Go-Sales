package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/repository"
	"github.com/jason127vip-dot/Go-Sales/response"
	"github.com/jason127vip-dot/Go-Sales/service"
)

type AIReviewHandler struct{ service *service.AIReviewService }

func NewAIReviewHandler(service *service.AIReviewService) *AIReviewHandler {
	return &AIReviewHandler{service: service}
}

func (h *AIReviewHandler) Review(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.AIReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid AI review request")
		return
	}
	result, err := h.service.Review(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAIQuestion) || errors.Is(err, repository.ErrOnlyDraftOrders) {
			response.Error(c, 400, err.Error())
			return
		}
		handleError(c, err)
		return
	}
	response.Success(c, result)
}
