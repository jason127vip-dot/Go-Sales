package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/response"
	"gorm.io/gorm"
	"strconv"
)

func BranchMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.FullPath() {
		case "/health", "/api/branches", "/api/branches/:id", "/api/customers", "/api/customers/:id", "/api/products", "/api/products/:id":
			c.Next()
			return
		}
		id, err := strconv.ParseUint(c.GetHeader("X-Branch-ID"), 10, 32)
		if err != nil || id == 0 {
			response.Error(c, 400, "select a branch before continuing")
			c.Abort()
			return
		}
		var branch model.Branch
		if err := db.WithContext(c.Request.Context()).First(&branch, uint(id)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				response.Error(c, 400, "selected branch does not exist")
			} else {
				response.Error(c, 500, "unable to load branch")
			}
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(model.WithBranch(c.Request.Context(), uint(id)))
		c.Next()
	}
}
