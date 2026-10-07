package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/config"
	"github.com/jason127vip-dot/Go-Sales/handler"
	"github.com/jason127vip-dot/Go-Sales/middleware"
	"github.com/jason127vip-dot/Go-Sales/repository"
	"github.com/jason127vip-dot/Go-Sales/router"
	"github.com/jason127vip-dot/Go-Sales/service"
)

func main() {
	cfg := config.LoadConfig()
	db, err := config.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := config.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.BranchMiddleware(db))
	branches := handler.NewBranchHandler(service.NewBranchService(repository.NewBranchRepository(db)))
	r.GET("/api/branches", branches.FindAll)
	r.POST("/api/branches", branches.Save)
	r.PUT("/api/branches/:id", branches.Save)
	credits := handler.NewCreditHandler(service.NewCreditService(repository.NewCreditRepository(db)))
	r.GET("/api/customer-credits", credits.Summary)
	r.PUT("/api/customer-credits/:id", credits.SetCustomerLimit)
	prices := handler.NewPriceListHandler(service.NewPriceListService(repository.NewPriceListRepository(db)))
	r.GET("/api/price-lists", prices.FindAll)
	r.POST("/api/price-lists", prices.Save)
	r.PUT("/api/price-lists/:id", prices.Save)
	r.DELETE("/api/price-lists/:id", prices.Delete)
	r.GET("/api/price-lists/resolve", prices.Resolve)
	customerRepository := repository.NewCustomerRepository(db)
	customerService := service.NewCustomerService(customerRepository)
	productRepository := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepository)
	salesOrderRepository := repository.NewSalesOrderRepository(db)
	salesOrderService := service.NewSalesOrderService(salesOrderRepository)
	documentRepository := repository.NewSalesDocumentRepository(db)
	documentService := service.NewSalesDocumentService(documentRepository)
	aiReviewService := service.NewAIReviewService(repository.NewOrderReviewRepository(db), service.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel))

	router.RegisterRoutes(
		r,
		handler.NewHealthHandler(),
		handler.NewCustomerHandler(customerService),
		handler.NewProductHandler(productService),
		handler.NewSalesOrderHandler(salesOrderService),
		handler.NewSalesDocumentHandler(documentService),
		handler.NewAIReviewHandler(aiReviewService),
	)

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
