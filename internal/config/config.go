package config

import (
	"os"
	"strings"
)

// Provider names match CLI and environment conventions.
const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderOllama    = "ollama"
)

// FromEnv loads provider, API key, model, and Ollama base URL from the environment.
func FromEnv() (provider, apiKey, model, ollamaBase string) {
	provider = strings.ToLower(strings.TrimSpace(os.Getenv("AIWK_PROVIDER")))
	if provider == "" {
		provider = ProviderOpenAI
	}
	apiKey = strings.TrimSpace(os.Getenv("AIWK_API_KEY"))
	model = strings.TrimSpace(os.Getenv("AIWK_MODEL"))
	ollamaBase = strings.TrimSpace(os.Getenv("AIWK_OLLAMA_HOST"))
	if ollamaBase == "" {
		ollamaBase = "http://127.0.0.1:11434"
	}
	return provider, apiKey, model, ollamaBase
}

// DefaultModel returns a sensible default per provider when AIWK_MODEL is unset.
func DefaultModel(provider string) string {
	switch strings.ToLower(provider) {
	case ProviderAnthropic:
		return "claude-3-5-sonnet-20241022"
	case ProviderOllama:
		return "llama3.2"
	default:
		return "gpt-4o-mini"
	}
}
