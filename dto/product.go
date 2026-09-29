package dto

type CreateProductRequest struct {
	ProductCode   string  `json:"productCode" binding:"required,max=50"`
	Barcode       string  `json:"barcode" binding:"max=100"`
	Name          string  `json:"name" binding:"required,max=200"`
	Specification string  `json:"specification" binding:"max=500"`
	Unit          string  `json:"unit" binding:"required,max=30"`
	UnitPrice     float64 `json:"unitPrice" binding:"gte=0"`
	Status        string  `json:"status" binding:"required,oneof=active inactive"`
}

type UpdateProductRequest = CreateProductRequest
