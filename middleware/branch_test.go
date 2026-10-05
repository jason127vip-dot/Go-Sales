package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBranchSelectionRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BranchMiddleware(nil))
	r.GET("/api/sales-orders", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/customers", func(c *gin.Context) { c.Status(200) })
	for _, header := range []string{"", "0", "bad", "-1"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/sales-orders", nil)
		req.Header.Set("X-Branch-ID", header)
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("header %q: status %d", header, w.Code)
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/customers", nil))
	if w.Code != 200 {
		t.Fatalf("shared customers require branch: %d", w.Code)
	}
}
