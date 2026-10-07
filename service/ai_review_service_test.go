package service

import "testing"

func TestAssessReview(t *testing.T) {
	tests := []struct {
		name       string
		credit     ReviewCredit
		aging      ReviewAging
		risk       string
		status     string
		canConfirm bool
	}{
		{"over limit", ReviewCredit{ControlEnabled: true, ExceededAmount: 25}, ReviewAging{}, "high", "over_limit", false},
		{"aged invoices", ReviewCredit{ControlEnabled: true}, ReviewAging{OlderThan30DaysCount: 1}, "medium", "within_limit", true},
		{"within limit", ReviewCredit{ControlEnabled: true}, ReviewAging{}, "low", "within_limit", true},
		{"disabled", ReviewCredit{ControlEnabled: false}, ReviewAging{}, "informational", "control_disabled", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := assessReview(test.credit, test.aging)
			if result.RiskLevel != test.risk || result.CreditStatus != test.status || result.CanConfirmUnderCreditRule != test.canConfirm {
				t.Fatalf("unexpected assessment: %+v", result)
			}
		})
	}
}

func TestMockReviewUsesQuestionLanguage(t *testing.T) {
	result := &AIReviewResult{
		Order:      ReviewOrder{OrderAmount: 5510},
		Credit:     ReviewCredit{CreditLimit: 5000, AvailableCredit: 5000, ProjectedExposure: 5510, ExceededAmount: 510},
		Assessment: ReviewAssessment{CreditStatus: "over_limit"},
	}
	review := mockReviewText(result, "你能说中文吗？")
	if !containsHan(review.Answer) || !containsHan(review.Recommendation) || !containsHan(review.SuggestedQuestions[0]) {
		t.Fatalf("expected Chinese review content: %+v", review)
	}
}
