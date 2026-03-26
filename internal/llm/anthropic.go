package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Anthropic calls the Messages API.
type Anthropic struct {
	HTTPClient *http.Client
	APIKey     string
	Model      string
	BaseURL    string
}

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	Messages  []anthropicMsg `json:"messages"`
	System    string         `json:"system"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (a *Anthropic) Complete(ctx context.Context, system, user string) (string, error) {
	if strings.TrimSpace(a.APIKey) == "" {
		return "", fmt.Errorf("anthropic: missing API key (set AIWK_API_KEY)")
	}
	base := strings.TrimSuffix(strings.TrimSpace(a.BaseURL), "/")
	if base == "" {
		base = "https://api.anthropic.com/v1"
	}
	url := base + "/messages"
	body, err := json.Marshal(anthropicRequest{
		Model:     a.Model,
		MaxTokens: 4096,
		System:    system,
		Messages: []anthropicMsg{
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	client := a.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var parsed anthropicResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("anthropic: decode: %w; body=%s", err, truncate(string(raw), 500))
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("anthropic: %s", parsed.Error.Message)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("anthropic: http %d: %s", res.StatusCode, truncate(string(raw), 500))
	}
	var out strings.Builder
	for _, b := range parsed.Content {
		if b.Type == "text" {
			out.WriteString(b.Text)
		}
	}
	if out.Len() == 0 {
		return "", fmt.Errorf("anthropic: empty content")
	}
	return out.String(), nil
}
