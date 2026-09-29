package dto

type CreateCustomerRequest struct {
	CustomerCode  string `json:"customerCode" binding:"required,max=50"`
	Name          string `json:"name" binding:"required,max=200"`
	ContactPerson string `json:"contactPerson" binding:"max=100"`
	Phone         string `json:"phone" binding:"max=50"`
	Email         string `json:"email" binding:"max=150"`
	Address       string `json:"address" binding:"max=500"`
	PaymentTerms  string `json:"paymentTerms" binding:"max=100"`
	Status        string `json:"status" binding:"required,oneof=active inactive"`
}

type UpdateCustomerRequest = CreateCustomerRequest
