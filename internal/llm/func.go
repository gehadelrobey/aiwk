package llm

import "context"

// Func adapts a function to the Client interface (useful in tests).
type Func func(ctx context.Context, system, user string) (string, error)

func (f Func) Complete(ctx context.Context, system, user string) (string, error) {
	return f(ctx, system, user)
}
