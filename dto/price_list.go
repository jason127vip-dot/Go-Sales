package dto

type SavePriceListRequest struct {
	CustomerID uint    `json:"customerId" binding:"required"`
	ProductID  uint    `json:"productId" binding:"required"`
	StartDate  string  `json:"startDate" binding:"required,datetime=2006-01-02"`
	EndDate    string  `json:"endDate" binding:"required,datetime=2006-01-02"`
	UnitPrice  float64 `json:"unitPrice" binding:"gte=0"`
}
