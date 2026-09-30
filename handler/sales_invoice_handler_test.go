package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jason127vip-dot/Go-Sales/repository"
)

func TestInvoiceAndPaymentRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &SalesDocumentHandler{}
	for _, test := range []struct {
		name, body string
		handle     gin.HandlerFunc
	}{
		{"missing outbound", `{"invoiceDate":"2026-09-30"}`, h.CreateInvoice},
		{"invalid date", `{"salesOutboundId":1,"invoiceDate":"2026-02-30"}`, h.CreateInvoice},
		{"oversized remarks", `{"salesOutboundId":1,"invoiceDate":"2026-09-30","remarks":"` + strings.Repeat("x", 1001) + `"}`, h.CreateInvoice},
		{"old order-only payload", `{"salesOrderId":1,"paymentDate":"2026-09-30","amount":5,"method":"Cash"}`, h.CreatePayment},
		{"zero amount", `{"salesInvoiceId":1,"paymentDate":"2026-09-30","amount":0,"method":"Cash"}`, h.CreatePayment},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/", test.handle)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestInvoiceBusinessErrorsAreClientErrors(t *testing.T) {
	for _, err := range []error{repository.ErrAlreadyInvoiced, repository.ErrDocumentInUse, repository.ErrLegacyPayment, repository.ErrInvoiceRequired, repository.ErrDocumentState} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		handleDocumentError(c, err)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%v: got %d", err, w.Code)
		}
	}
}
