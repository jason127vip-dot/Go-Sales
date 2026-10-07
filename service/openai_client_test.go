package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIClientReview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("missing authorization header")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["store"] != false || body["model"] != "test-model" {
			t.Fatalf("unexpected request: %+v", body)
		}
		instructions, _ := body["instructions"].(string)
		if !strings.Contains(instructions, "same language") {
			t.Fatalf("missing response language instruction")
		}
		text, ok := body["text"].(map[string]any)
		if !ok {
			t.Fatalf("missing text configuration")
		}
		format, ok := text["format"].(map[string]any)
		if !ok || format["type"] != "json_schema" {
			t.Fatalf("missing structured output format")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"{\"summary\":\"Summary\",\"answer\":\"Answer\",\"recommendation\":\"Recommendation\",\"suggestedQuestions\":[\"Question\"]}"}]}]}`))
	}))
	defer server.Close()

	client := &OpenAIClient{apiKey: "test-key", model: "test-model", url: server.URL, client: server.Client()}
	review, err := client.Review(context.Background(), AIReviewPrompt{})
	if err != nil {
		t.Fatal(err)
	}
	if review.Summary != "Summary" || review.Answer != "Answer" || len(review.SuggestedQuestions) != 1 {
		t.Fatalf("unexpected review: %+v", review)
	}
}
