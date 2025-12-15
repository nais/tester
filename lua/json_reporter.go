package lua

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nais/tester/lua/reporter"
)

type JSONReporter struct {
	file   string
	name   string
	runner string
	w      *json.Encoder
}

func NewJSONReporter(w io.Writer) *JSONReporter {
	return &JSONReporter{w: json.NewEncoder(w)}
}

func (r *JSONReporter) RunFile(ctx context.Context, filename string, fn func(reporter.Reporter)) {
	_ = r.w.Encode(map[string]any{
		"file":   filename,
		"action": "start",
	})
	fn(&JSONReporter{w: r.w, file: filename})
	_ = r.w.Encode(map[string]any{
		"file":   filename,
		"action": "end",
	})
}

func (r *JSONReporter) RunTest(ctx context.Context, runner, name string, fn func(reporter.Reporter)) {
	_ = r.w.Encode(map[string]any{
		"file":   r.file,
		"name":   name,
		"runner": runner,
		"action": "start",
	})

	fn(&JSONReporter{w: r.w, file: r.file, name: name, runner: runner})

	_ = r.w.Encode(map[string]any{
		"file":   r.file,
		"name":   name,
		"runner": runner,
		"action": "end",
	})
}

func (r *JSONReporter) Error(msg string, args ...any) {
	_ = r.w.Encode(map[string]any{
		"error":  fmt.Sprintf(msg, args...),
		"file":   r.file,
		"name":   r.name,
		"runner": r.runner,
	})
}

func (r *JSONReporter) Info(info reporter.Info) {
	_ = r.w.Encode(map[string]any{
		"info":   info,
		"file":   r.file,
		"name":   r.name,
		"runner": r.runner,
	})
}
