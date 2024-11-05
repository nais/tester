package reporter

import "context"

type Reporter interface {
	RunFile(ctx context.Context, filename string, fn func(Reporter))
	RunTest(ctx context.Context, runner, name string, fn func(Reporter))
	Error(msg string, args ...any)
}
