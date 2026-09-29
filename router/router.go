package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/handler"
)

func RegisterRoutes(r *gin.Engine, healthHandler *handler.HealthHandler, customerHandler *handler.CustomerHandler, productHandler *handler.ProductHandler) {
	r.GET("/health", healthHandler.Check)

	api := r.Group("/api")
	customers := api.Group("/customers")
	customers.GET("", customerHandler.FindAll)
	customers.POST("", customerHandler.Create)
	customers.PUT("/:id", customerHandler.Update)
	customers.DELETE("/:id", customerHandler.Delete)

	products := api.Group("/products")
	products.GET("", productHandler.FindAll)
	products.POST("", productHandler.Create)
	products.PUT("/:id", productHandler.Update)
	products.DELETE("/:id", productHandler.Delete)
}
