package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok"})
}
