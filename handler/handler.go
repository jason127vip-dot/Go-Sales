package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/response"
	"gorm.io/gorm"
)

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "invalid id")
		return 0, false
	}
	return uint(id), true
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "record not found")
		return
	}
	response.Error(c, 500, err.Error())
}
