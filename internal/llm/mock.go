package llm

import (
	"context"
	"fmt"
)

// Mock returns a fixed response (for tests).
type Mock struct {
	Text string
	Err  error
}

func (m *Mock) Complete(ctx context.Context, system, user string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if m.Text == "" {
		return "", fmt.Errorf("mock: empty Text")
	}
	return m.Text, nil
}
