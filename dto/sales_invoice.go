package dto

type SaveSalesInvoiceRequest struct {
	SalesOutboundID uint   `json:"salesOutboundId" binding:"required"`
	InvoiceDate     string `json:"invoiceDate" binding:"required,datetime=2006-01-02"`
	Remarks         string `json:"remarks" binding:"max=1000"`
}
