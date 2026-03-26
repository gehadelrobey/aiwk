package llm

import (
	"context"
)

// Client generates awk source from a natural-language task.
type Client interface {
	Complete(ctx context.Context, system, user string) (string, error)
}
