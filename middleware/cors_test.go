package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSPreflightStopsBeforeBranchValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware(), BranchMiddleware(nil))
	r.GET("/api/branches", func(c *gin.Context) { c.Status(200) })
	r.GET("/api/sales-orders", func(c *gin.Context) { c.Status(200) })
	for _, path := range []string{"/api/branches", "/api/sales-orders"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, path, nil)
		req.Header.Set("Origin", "http://192.168.69.23:5173")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "content-type,x-branch-id")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent || w.Body.Len() != 0 {
			t.Fatalf("%s: expected empty 204, got %d %s", path, w.Code, w.Body.String())
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "http://192.168.69.23:5173" || !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "X-Branch-ID") {
			t.Fatalf("%s: missing CORS headers: %v", path, w.Header())
		}
	}
}
