package repository

import (
	"testing"
	"time"
)

func TestDocumentNumberBaseIncludesBranchAndYear(t *testing.T) {
	date := time.Date(2026, time.October, 6, 0, 0, 0, 0, time.UTC)
	tests := map[string]string{
		"SO":  "SO-AKL-2026-",
		"OUT": "OUT-AKL-2026-",
		"INV": "INV-AKL-2026-",
		"PAY": "PAY-AKL-2026-",
	}
	for prefix, expected := range tests {
		if actual := documentNumberBase(prefix, "AKL", date); actual != expected {
			t.Fatalf("%s: expected %q, got %q", prefix, expected, actual)
		}
	}
}
