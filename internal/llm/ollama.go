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

// Ollama calls the local /api/generate endpoint.
type Ollama struct {
	HTTPClient *http.Client
	Model      string
	BaseURL    string
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

func (o *Ollama) Complete(ctx context.Context, system, user string) (string, error) {
	base := strings.TrimSuffix(strings.TrimSpace(o.BaseURL), "/")
	if base == "" {
		base = "http://127.0.0.1:11434"
	}
	url := base + "/api/generate"
	prompt := "SYSTEM:\n" + system + "\n\nUSER:\n" + user + "\n"
	body, err := json.Marshal(ollamaRequest{
		Model:  o.Model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := o.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 300 * time.Second}
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
	var parsed ollamaResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("ollama: decode: %w; body=%s", err, truncate(string(raw), 500))
	}
	if parsed.Error != "" {
		return "", fmt.Errorf("ollama: %s", parsed.Error)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("ollama: http %d: %s", res.StatusCode, truncate(string(raw), 500))
	}
	if strings.TrimSpace(parsed.Response) == "" {
		return "", fmt.Errorf("ollama: empty response")
	}
	return parsed.Response, nil
}
