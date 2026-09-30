package dto

type SalesOutboundLineRequest struct {
	SalesOrderLineID uint    `json:"salesOrderLineId" binding:"required"`
	OutboundQuantity float64 `json:"outboundQuantity" binding:"gt=0"`
}

type CreateSalesOutboundRequest struct {
	SalesOrderID uint                       `json:"salesOrderId" binding:"required"`
	OutboundDate string                     `json:"outboundDate" binding:"required,datetime=2006-01-02"`
	Lines        []SalesOutboundLineRequest `json:"lines" binding:"required,min=1,dive"`
}

type CreatePaymentRequest struct {
	SalesInvoiceID uint    `json:"salesInvoiceId" binding:"required"`
	PaymentDate    string  `json:"paymentDate" binding:"required,datetime=2006-01-02"`
	Amount         float64 `json:"amount" binding:"gt=0"`
	Method         string  `json:"method" binding:"required,max=50"`
	ReferenceNo    string  `json:"referenceNo" binding:"max=100"`
}
