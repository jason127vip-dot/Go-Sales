package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/handler"
)

func RegisterRoutes(r *gin.Engine, healthHandler *handler.HealthHandler, customerHandler *handler.CustomerHandler, productHandler *handler.ProductHandler, salesOrderHandler *handler.SalesOrderHandler, documentHandler *handler.SalesDocumentHandler, aiReviewHandler *handler.AIReviewHandler) {
	r.GET("/health", healthHandler.Check)

	api := r.Group("/api")
	api.GET("/dashboard", documentHandler.Dashboard)
	api.GET("/reports/sales-order-payments", documentHandler.PaymentReport)
	api.GET("/reports/ar-aging", documentHandler.ARAgingReport)
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

	orders := api.Group("/sales-orders")
	orders.GET("", salesOrderHandler.FindAll)
	orders.POST("", salesOrderHandler.Create)
	orders.PUT("/:id", salesOrderHandler.Update)
	orders.POST("/:id/confirm", salesOrderHandler.Confirm)
	orders.POST("/:id/cancel-confirmation", salesOrderHandler.CancelConfirmation)
	orders.POST("/:id/ai-review", aiReviewHandler.Review)
	orders.DELETE("/:id", salesOrderHandler.Delete)
	orders.GET("/:id/execution", documentHandler.Execution)

	outbounds := api.Group("/sales-outbounds")
	outbounds.GET("", documentHandler.FindOutbounds)
	outbounds.GET("/available-orders", documentHandler.AvailableOrders)
	outbounds.GET("/order/:id/lines", documentHandler.OutboundLines)
	outbounds.POST("", documentHandler.CreateOutbound)
	outbounds.PUT("/:id", documentHandler.UpdateOutbound)
	outbounds.POST("/:id/confirm", documentHandler.ConfirmOutbound)
	outbounds.POST("/:id/cancel-confirmation", documentHandler.CancelOutbound)
	outbounds.DELETE("/:id", documentHandler.DeleteOutbound)

	invoices := api.Group("/sales-invoices")
	invoices.GET("", documentHandler.FindInvoices)
	invoices.GET("/available-outbounds", documentHandler.InvoiceOutbounds)
	invoices.POST("", documentHandler.CreateInvoice)
	invoices.PUT("/:id", documentHandler.UpdateInvoice)
	invoices.POST("/:id/confirm", documentHandler.ConfirmInvoice)
	invoices.POST("/:id/cancel-confirmation", documentHandler.CancelInvoice)
	invoices.DELETE("/:id", documentHandler.DeleteInvoice)

	payments := api.Group("/payments")
	payments.GET("", documentHandler.FindPayments)
	payments.GET("/invoice-summaries", documentHandler.PaymentSummaries)
	payments.POST("", documentHandler.CreatePayment)
	payments.PUT("/:id", documentHandler.UpdatePayment)
	payments.POST("/:id/confirm", documentHandler.ConfirmPayment)
	payments.POST("/:id/cancel-confirmation", documentHandler.CancelPayment)
	payments.DELETE("/:id", documentHandler.DeletePayment)
}
