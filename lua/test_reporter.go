package lua

import (
	"context"
	"fmt"
	"testing"

	"github.com/nais/tester/lua/reporter"
)

type TestReporter struct {
	t    *testing.T
	name string
}

func NewTestReporter(t *testing.T) *TestReporter {
	return &TestReporter{t: t}
}

func (r *TestReporter) RunFile(ctx context.Context, filename string, fn func(reporter.Reporter)) {
	r.t.Run(filename, func(t *testing.T) {
		fn(&TestReporter{t: t, name: filename})
	})
}

func (r *TestReporter) RunTest(ctx context.Context, runner string, name string, fn func(reporter.Reporter)) {
	r.t.Run(name, func(t *testing.T) {
		fn(&TestReporter{t: t, name: r.name + "////" + name})
	})
}

func (r *TestReporter) Error(msg string, args ...any) {
	r.t.Errorf(msg, args...)
}

func (r *TestReporter) Info(info reporter.Info) {
	r.t.Logf("[%s] %s: %s", info.Type, info.Title, fmt.Sprintf("%.100s", info.Content))
}
