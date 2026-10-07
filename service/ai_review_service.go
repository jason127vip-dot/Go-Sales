package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/jason127vip-dot/Go-Sales/dto"
	"github.com/jason127vip-dot/Go-Sales/repository"
)

var ErrInvalidAIQuestion = errors.New("question must be 2000 characters or fewer")

type AIReviewRepository interface {
	Context(context.Context, uint) (*repository.OrderReviewContext, error)
}

type AIReviewClient interface {
	Review(context.Context, AIReviewPrompt) (AIReviewText, error)
}

type AIReviewService struct {
	repository AIReviewRepository
	client     AIReviewClient
}

type AIReviewPrompt struct {
	Context    repository.OrderReviewContext `json:"context"`
	Assessment ReviewAssessment              `json:"assessment"`
	Question   string                        `json:"question,omitempty"`
	History    []dto.AIReviewMessage         `json:"history,omitempty"`
}

type ReviewOrder struct {
	OrderNo      string  `json:"orderNo"`
	CustomerCode string  `json:"customerCode"`
	CustomerName string  `json:"customerName"`
	OrderAmount  float64 `json:"orderAmount"`
}

type ReviewCredit struct {
	ControlEnabled    bool    `json:"controlEnabled"`
	Configured        bool    `json:"configured"`
	CreditLimit       float64 `json:"creditLimit"`
	UsedCredit        float64 `json:"usedCredit"`
	AvailableCredit   float64 `json:"availableCredit"`
	ProjectedExposure float64 `json:"projectedExposure"`
	ExceededAmount    float64 `json:"exceededAmount"`
}

type ReviewAging struct {
	OutstandingAmount     float64                    `json:"outstandingAmount"`
	OlderThan30DaysAmount float64                    `json:"olderThan30DaysAmount"`
	OlderThan30DaysCount  int                        `json:"olderThan30DaysCount"`
	OldestAgingDays       int                        `json:"oldestAgingDays"`
	Invoices              []repository.ReviewInvoice `json:"invoices"`
}

type ReviewAssessment struct {
	RiskLevel                 string   `json:"riskLevel"`
	CreditStatus              string   `json:"creditStatus"`
	CanConfirmUnderCreditRule bool     `json:"canConfirmUnderCreditRule"`
	Reasons                   []string `json:"reasons"`
}

type AIReviewText struct {
	Summary            string   `json:"summary"`
	Answer             string   `json:"answer"`
	Recommendation     string   `json:"recommendation"`
	SuggestedQuestions []string `json:"suggestedQuestions"`
}

type AIReviewResult struct {
	Order         ReviewOrder      `json:"order"`
	Credit        ReviewCredit     `json:"credit"`
	Aging         ReviewAging      `json:"aging"`
	Assessment    ReviewAssessment `json:"assessment"`
	AI            AIReviewText     `json:"ai"`
	Source        string           `json:"source"`
	AIUnavailable bool             `json:"aiUnavailable"`
}

func NewAIReviewService(repository AIReviewRepository, client AIReviewClient) *AIReviewService {
	return &AIReviewService{repository: repository, client: client}
}

func (s *AIReviewService) Review(ctx context.Context, id uint, req dto.AIReviewRequest) (*AIReviewResult, error) {
	question := strings.TrimSpace(req.Question)
	if len(question) > 2000 {
		return nil, ErrInvalidAIQuestion
	}
	history := sanitizeHistory(req.History)
	reviewContext, err := s.repository.Context(ctx, id)
	if err != nil {
		return nil, err
	}
	result := buildReviewResult(reviewContext)
	prompt := AIReviewPrompt{Context: *reviewContext, Assessment: result.Assessment, Question: question, History: history}
	result.AI = mockReviewText(result, question)
	result.Source = "mock"
	if s.client != nil {
		ai, clientErr := s.client.Review(ctx, prompt)
		if clientErr == nil {
			result.AI = ai
			result.Source = "openai"
		} else {
			result.AIUnavailable = true
		}
	}
	return result, nil
}

func sanitizeHistory(messages []dto.AIReviewMessage) []dto.AIReviewMessage {
	if len(messages) > 6 {
		messages = messages[len(messages)-6:]
	}
	result := make([]dto.AIReviewMessage, 0, len(messages))
	for _, message := range messages {
		content := strings.TrimSpace(message.Content)
		if (message.Role != "user" && message.Role != "assistant") || content == "" {
			continue
		}
		if len(content) > 2000 {
			content = content[:2000]
		}
		result = append(result, dto.AIReviewMessage{Role: message.Role, Content: content})
	}
	return result
}

func buildReviewResult(row *repository.OrderReviewContext) *AIReviewResult {
	credit := ReviewCredit{
		ControlEnabled: row.CreditControlEnabled, Configured: row.CreditConfigured,
		CreditLimit: row.CreditLimit, UsedCredit: row.UsedCredit, AvailableCredit: row.AvailableCredit,
		ProjectedExposure: roundReviewMoney(row.UsedCredit + row.OrderAmount),
		ExceededAmount:    roundReviewMoney(math.Max(0, row.OrderAmount-row.AvailableCredit)),
	}
	aging := ReviewAging{Invoices: row.Invoices}
	for _, invoice := range row.Invoices {
		aging.OutstandingAmount += invoice.OutstandingAmount
		if invoice.AgingDays > aging.OldestAgingDays {
			aging.OldestAgingDays = invoice.AgingDays
		}
		if invoice.AgingDays > 30 {
			aging.OlderThan30DaysCount++
			aging.OlderThan30DaysAmount += invoice.OutstandingAmount
		}
	}
	aging.OutstandingAmount = roundReviewMoney(aging.OutstandingAmount)
	aging.OlderThan30DaysAmount = roundReviewMoney(aging.OlderThan30DaysAmount)
	assessment := assessReview(credit, aging)
	return &AIReviewResult{
		Order:  ReviewOrder{OrderNo: row.OrderNo, CustomerCode: row.CustomerCode, CustomerName: row.CustomerName, OrderAmount: row.OrderAmount},
		Credit: credit, Aging: aging, Assessment: assessment,
	}
}

func assessReview(credit ReviewCredit, aging ReviewAging) ReviewAssessment {
	assessment := ReviewAssessment{RiskLevel: "low", CreditStatus: "within_limit", CanConfirmUnderCreditRule: true, Reasons: make([]string, 0)}
	if !credit.ControlEnabled {
		assessment.RiskLevel = "informational"
		assessment.CreditStatus = "control_disabled"
		assessment.Reasons = append(assessment.Reasons, "Credit control is disabled for this branch")
	} else if credit.ExceededAmount > 0 {
		assessment.RiskLevel = "high"
		assessment.CreditStatus = "over_limit"
		assessment.CanConfirmUnderCreditRule = false
		assessment.Reasons = append(assessment.Reasons, fmt.Sprintf("Order amount exceeds available credit by %s", money(credit.ExceededAmount)))
	} else {
		assessment.Reasons = append(assessment.Reasons, "Order amount is within the available credit")
	}
	if aging.OlderThan30DaysCount > 0 {
		if assessment.RiskLevel == "low" {
			assessment.RiskLevel = "medium"
		}
		assessment.Reasons = append(assessment.Reasons, fmt.Sprintf("Customer has %d outstanding invoice(s) older than 30 days", aging.OlderThan30DaysCount))
	}
	return assessment
}

func mockReviewText(result *AIReviewResult, question string) AIReviewText {
	if containsHan(question) {
		return mockChineseReviewText(result, question)
	}
	text := AIReviewText{SuggestedQuestions: []string{"Which invoices are outstanding?", "How much must the customer pay?", "How is available credit calculated?"}}
	switch result.Assessment.CreditStatus {
	case "over_limit":
		text.Summary = fmt.Sprintf("This order is %s above the customer's available credit.", money(result.Credit.ExceededAmount))
		text.Recommendation = "Receive sufficient payment or revise the order before attempting confirmation."
	case "control_disabled":
		text.Summary = "Credit usage is available for reference, but credit control is disabled for this branch."
		text.Recommendation = "Review the customer's outstanding balance before confirming the order."
	default:
		text.Summary = "This order is within the customer's available credit."
		text.Recommendation = "The order can proceed under the current credit rule."
	}
	if result.Aging.OlderThan30DaysCount > 0 {
		text.Summary += fmt.Sprintf(" The customer has %d outstanding invoice(s) older than 30 days.", result.Aging.OlderThan30DaysCount)
		if result.Assessment.CreditStatus != "over_limit" {
			text.Recommendation = "Review the aged outstanding invoices before confirming the order."
		}
	}
	text.Answer = text.Summary
	lower := strings.ToLower(question)
	if question == "" {
		return text
	}
	switch {
	case strings.Contains(lower, "invoice") || strings.Contains(lower, "outstanding"):
		if len(result.Aging.Invoices) == 0 {
			text.Answer = "The customer has no outstanding confirmed invoices in this branch."
		} else {
			parts := make([]string, 0, len(result.Aging.Invoices))
			for _, invoice := range result.Aging.Invoices {
				parts = append(parts, fmt.Sprintf("%s: %s outstanding, %d days old", invoice.InvoiceNo, money(invoice.OutstandingAmount), invoice.AgingDays))
			}
			text.Answer = strings.Join(parts, "; ") + "."
		}
	case strings.Contains(lower, "pay") || strings.Contains(lower, "much"):
		if result.Credit.ExceededAmount > 0 {
			text.Answer = fmt.Sprintf("At least %s of confirmed payment is needed to bring this order within the current available credit.", money(result.Credit.ExceededAmount))
		} else {
			text.Answer = "No payment is required for this order to remain within the current available credit."
		}
	case strings.Contains(lower, "calculat") || strings.Contains(lower, "available credit"):
		text.Answer = fmt.Sprintf("Available credit is credit limit %s minus used credit %s, which equals %s.", money(result.Credit.CreditLimit), money(result.Credit.UsedCredit), money(result.Credit.AvailableCredit))
	}
	return text
}

func mockChineseReviewText(result *AIReviewResult, question string) AIReviewText {
	text := AIReviewText{SuggestedQuestions: []string{"哪些发票尚未结清？", "客户至少需要付款多少？", "可用信用额度是如何计算的？"}}
	switch result.Assessment.CreditStatus {
	case "over_limit":
		text.Summary = fmt.Sprintf("该订单超出客户可用信用额度 %s。", money(result.Credit.ExceededAmount))
		text.Recommendation = "收到足够的客户付款或调整订单金额后，再尝试确认订单。"
	case "control_disabled":
		text.Summary = "当前分支未启用信用控制，信用数据仅供参考。"
		text.Recommendation = "确认订单前，请检查客户的未结余额。"
	default:
		text.Summary = "该订单金额在客户当前可用信用额度以内。"
		text.Recommendation = "按照当前信用规则，该订单可以继续处理。"
	}
	if result.Aging.OlderThan30DaysCount > 0 {
		text.Summary += fmt.Sprintf(" 客户还有 %d 张账龄超过 30 天的未结发票。", result.Aging.OlderThan30DaysCount)
		if result.Assessment.CreditStatus != "over_limit" {
			text.Recommendation = "确认订单前，请先检查账龄较长的未结发票。"
		}
	}
	text.Answer = text.Summary
	switch {
	case strings.Contains(question, "中文"):
		text.Answer = "可以，我会使用中文回答你后续的信用问题。" + text.Summary
	case strings.Contains(question, "可用额度") || strings.Contains(question, "信用额度") || strings.Contains(question, "计算") || strings.Contains(question, "怎么算"):
		text.Answer = fmt.Sprintf("客户信用额度为 %s，已使用额度为 %s，因此当前可用额度为 %s。加入这张 %s 的订单后，预计信用占用为 %s。", money(result.Credit.CreditLimit), money(result.Credit.UsedCredit), money(result.Credit.AvailableCredit), money(result.Order.OrderAmount), money(result.Credit.ProjectedExposure))
	case strings.Contains(question, "发票") || strings.Contains(question, "未结") || strings.Contains(question, "逾期"):
		if len(result.Aging.Invoices) == 0 {
			text.Answer = "该客户在当前分支没有已确认且尚未结清的发票。"
		} else {
			parts := make([]string, 0, len(result.Aging.Invoices))
			for _, invoice := range result.Aging.Invoices {
				parts = append(parts, fmt.Sprintf("%s：未结金额 %s，账龄 %d 天", invoice.InvoiceNo, money(invoice.OutstandingAmount), invoice.AgingDays))
			}
			text.Answer = strings.Join(parts, "；") + "。"
		}
	case strings.Contains(question, "付款") || strings.Contains(question, "支付") || strings.Contains(question, "需要付"):
		if result.Credit.ExceededAmount > 0 {
			text.Answer = fmt.Sprintf("至少需要确认 %s 的客户付款，才能让这张订单回到当前可用信用额度以内。", money(result.Credit.ExceededAmount))
		} else {
			text.Answer = "这张订单目前没有因为信用额度而要求客户先付款。"
		}
	}
	return text
}

func containsHan(value string) bool {
	for _, character := range value {
		if unicode.Is(unicode.Han, character) {
			return true
		}
	}
	return false
}

func money(value float64) string             { return fmt.Sprintf("$%.2f", value) }
func roundReviewMoney(value float64) float64 { return math.Round(value*100) / 100 }
