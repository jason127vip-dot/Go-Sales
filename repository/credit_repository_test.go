package repository

import (
	"errors"
	"strings"
	"testing"
)

func TestEvaluateCredit(t *testing.T) {
	position := creditPosition{Enabled: true, Limit: 100, Used: 80, Available: 20}

	warning, err := evaluateCredit(position, 20, false)
	if err != nil || warning != "" {
		t.Fatalf("order within available credit was rejected: warning=%q err=%v", warning, err)
	}

	warning, err = evaluateCredit(position, 30, false)
	if err != nil || !strings.Contains(warning, "exceeded by $10.00") {
		t.Fatalf("draft should return an over-limit warning: warning=%q err=%v", warning, err)
	}

	warning, err = evaluateCredit(position, 30, true)
	var creditError *CreditLimitExceededError
	if warning != "" || !errors.As(err, &creditError) || creditError.Available != 20 || creditError.Order != 30 {
		t.Fatalf("confirmation should be blocked with credit details: warning=%q err=%v", warning, err)
	}
}

func TestEvaluateCreditDoesNotBlockWhenDisabled(t *testing.T) {
	warning, err := evaluateCredit(creditPosition{Enabled: false, Available: 0}, 100, true)
	if err != nil || warning != "" {
		t.Fatalf("disabled credit control blocked the order: warning=%q err=%v", warning, err)
	}
}
