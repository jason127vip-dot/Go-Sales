package service

import (
	"math"
	"testing"
)

func TestValidCreditAmount(t *testing.T) {
	for _, value := range []float64{0, 10, 10.25, 999999.99} {
		if !validCreditAmount(value) {
			t.Fatalf("valid amount rejected: %v", value)
		}
	}
	for _, value := range []float64{-1, 0.001, 10.999, math.NaN(), math.Inf(1)} {
		if validCreditAmount(value) {
			t.Fatalf("invalid amount accepted: %v", value)
		}
	}
}
