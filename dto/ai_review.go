package dto

type AIReviewMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIReviewRequest struct {
	Question string            `json:"question"`
	History  []AIReviewMessage `json:"history"`
}
