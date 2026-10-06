package repository

import (
	"testing"
	"time"
)

func TestInvoiceAgingDaysAndBuckets(t *testing.T) {
	analysisDate := time.Date(2026, time.October, 6, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		invoiceDate time.Time
		days        int
		bucket      string
	}{
		{analysisDate.AddDate(0, 0, 1), 0, "0-30 Days"},
		{analysisDate, 0, "0-30 Days"},
		{analysisDate.AddDate(0, 0, -30), 30, "0-30 Days"},
		{analysisDate.AddDate(0, 0, -31), 31, "31-60 Days"},
		{analysisDate.AddDate(0, 0, -61), 61, "61-90 Days"},
		{analysisDate.AddDate(0, 0, -91), 91, "91-120 Days"},
		{analysisDate.AddDate(0, 0, -121), 121, "120+ Days"},
	}
	for _, test := range tests {
		days := invoiceAgingDays(test.invoiceDate, analysisDate)
		if days != test.days || agingBucket(days) != test.bucket {
			t.Fatalf("expected %d/%s, got %d/%s", test.days, test.bucket, days, agingBucket(days))
		}
	}
}
