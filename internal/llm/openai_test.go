package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAI_Complete(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(openAIResponse{
			Choices: []openAIChoice{{Message: openAIMessage{Content: "{ print $1 }"}}},
		})
	}))
	t.Cleanup(srv.Close)

	c := &OpenAI{
		APIKey:  "k",
		Model:   "m",
		BaseURL: srv.URL + "/v1",
	}
	got, err := c.Complete(context.Background(), "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if got != "{ print $1 }" {
		t.Fatalf("got %q", got)
	}
}
