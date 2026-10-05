package service

import (
	"math"
	"testing"
)

func TestValidPrice(t *testing.T) {
	for _, value := range []float64{0, 10, 10.25, 999999.99} {
		if !validPrice(value) {
			t.Fatalf("valid price rejected: %v", value)
		}
	}
	for _, value := range []float64{-1, 0.001, math.NaN(), math.Inf(1)} {
		if validPrice(value) {
			t.Fatalf("invalid price accepted: %v", value)
		}
	}
}
