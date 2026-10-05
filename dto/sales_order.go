package dto

type SalesOrderLineRequest struct {
	ProductID uint     `json:"productId" binding:"required"`
	UnitPrice *float64 `json:"unitPrice" binding:"omitempty,gte=0"`
	Quantity  float64  `json:"quantity" binding:"gt=0"`
}

type CreateSalesOrderRequest struct {
	CustomerID           uint                    `json:"customerId" binding:"required"`
	OrderDate            string                  `json:"orderDate" binding:"required,datetime=2006-01-02"`
	CustomerPONo         string                  `json:"customerPoNo" binding:"max=100"`
	ExpectedOutboundDate string                  `json:"expectedOutboundDate"`
	Salesperson          string                  `json:"salesperson" binding:"max=100"`
	Remarks              string                  `json:"remarks" binding:"max=1000"`
	Lines                []SalesOrderLineRequest `json:"lines" binding:"required,min=1,dive"`
}

type UpdateSalesOrderRequest = CreateSalesOrderRequest
