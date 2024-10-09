package lua

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/nais/tester/lua/spec"
)

type SetupFunc func(ctx context.Context) (runners []spec.Runner, close func(), err error)

type Manager struct {
	runners     []spec.Runner
	newConfigFn func() any
	setup       SetupFunc
}

func New(newConfigFn func() any, setup SetupFunc, runners ...spec.Runner) (*Manager, error) {
	if setup == nil {
		return nil, fmt.Errorf("setup function must be provided")
	}

	// Make sure newConfigFn returns nil or a pointer to a struct
	if newConfigFn != nil {
		ret := newConfigFn()
		if ret != nil {
			t := reflect.TypeOf(ret)
			if t.Kind() != reflect.Ptr {
				return nil, fmt.Errorf("newConfigFn must return a *pointer* to a struct")
			}
			if t.Elem().Kind() != reflect.Struct {
				return nil, fmt.Errorf("newConfigFn must return a pointer to a *struct*")
			}
		}
	}

	return &Manager{
		setup:       setup,
		newConfigFn: newConfigFn,
		runners:     runners,
	}, nil
}

type Reporter interface {
	RunFile(ctx context.Context, filename string, fn func(Reporter))
	RunTest(ctx context.Context, runner, name string, fn func(Reporter))
	Error(msg string, args ...any)
}

func (m *Manager) Run(ctx context.Context, dir string, reporter Reporter) error {
	runners, close, err := m.setup(ctx)
	if err != nil {
		return fmt.Errorf("setting up test environment: %w", err)
	}
	defer close()

	entries, err := filepath.Glob(filepath.Join(dir, "*.lua"))
	if err != nil {
		return fmt.Errorf("reading fs directory: %w", err)
	}

	for _, f := range entries {
		if filepath.Base(f) == specFilename {
			continue
		}

		reporter.RunFile(ctx, f, func(r Reporter) {
			s := newSuite(m, runners, r)
			s.run(ctx, f)
		})
	}

	return nil
}

func (m *Manager) GenerateSpec(dir string) error {
	path := filepath.Join(dir, specFilename)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", path, err)
	}
	defer f.Close()

	GenerateSpec(f, m.runners, m.newConfigFn())
	return nil
}
