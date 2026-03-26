package config

import "testing"

func TestDefaultModel(t *testing.T) {
	t.Parallel()
	if g := DefaultModel(ProviderOpenAI); g == "" {
		t.Fatal("empty openai default")
	}
	if g := DefaultModel(ProviderAnthropic); g == "" {
		t.Fatal("empty anthropic default")
	}
	if g := DefaultModel(ProviderOllama); g == "" {
		t.Fatal("empty ollama default")
	}
}
