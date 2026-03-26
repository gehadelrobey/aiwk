package llm

import (
	"fmt"
	"os"
	"strings"

	"github.com/gehadelrobey/aiwk/internal/config"
)

// NewFromConfig picks a client based on provider name and environment.
func NewFromConfig(provider, apiKey, model, ollamaBase string) (Client, error) {
	p := strings.ToLower(strings.TrimSpace(provider))
	if p == "" {
		p = config.ProviderOpenAI
	}
	if model == "" {
		model = config.DefaultModel(p)
	}
	switch p {
	case config.ProviderOpenAI:
		return &OpenAI{
			APIKey:  apiKey,
			Model:   model,
			BaseURL: strings.TrimSpace(os.Getenv("AIWK_OPENAI_BASE_URL")),
		}, nil
	case config.ProviderAnthropic:
		return &Anthropic{
			APIKey:  apiKey,
			Model:   model,
			BaseURL: strings.TrimSpace(os.Getenv("AIWK_ANTHROPIC_BASE_URL")),
		}, nil
	case config.ProviderOllama:
		return &Ollama{
			Model:   model,
			BaseURL: ollamaBase,
		}, nil
	default:
		return nil, fmt.Errorf("unknown provider %q (use openai, anthropic, ollama)", provider)
	}
}
