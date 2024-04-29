package testmanager

import "context"

type contextKey int

const (
	contextKeyTestDir contextKey = iota
)

// withTestDir returns a new context with the test directory name set.
func withTestDir(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextKeyTestDir, name)
}

// TestDir returns the test directory name from the context.
func TestDir(ctx context.Context) string {
	r, ok := ctx.Value(contextKeyTestDir).(string)
	if !ok {
		return ""
	}
	return r
}
