package lua

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/nais/tester/internal/webui"
	"github.com/nais/tester/lua/reporter"
	"github.com/nais/tester/lua/spec"
	"golang.org/x/sync/errgroup"
)

type SetupFunc func(ctx context.Context, dir string, config any) (runners []spec.Runner, close func(), err error)

type Manager struct {
	runners     []spec.Runner
	newConfigFn func() any
	setup       SetupFunc
	dir         string
	helpers     []*spec.Function
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

func (m *Manager) Run(ctx context.Context, dir string, report reporter.Reporter) error {
	m.dir = dir
	return m.run(ctx, report)
}

func (m *Manager) run(ctx context.Context, report reporter.Reporter) error {
	entries, err := filepath.Glob(filepath.Join(m.dir, "*.lua"))
	if err != nil {
		return fmt.Errorf("reading fs directory: %w", err)
	}

	for _, f := range entries {
		if filepath.Base(f) == specFilename {
			continue
		}

		report.RunFile(ctx, f, func(r reporter.Reporter) {
			s := newSuite(m, r)
			s.run(ctx, f)
		})
	}

	return nil
}

func (m *Manager) RunUI(ctx context.Context, dir string) error {
	m.dir = dir

	reporter := webui.NewSSEReporter(dir)

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		err := m.watch(ctx, dir, reporter)
		if err != nil {
			fmt.Println("WATCH ERROR", err)
		}
		return err
	})

	wg.Go(func() error {
		err := m.run(ctx, reporter)
		if err != nil {
			fmt.Println("RUN ERROR", err)
		}
		return err
	})

	wg.Go(func() error {
		err := webui.Run(ctx, reporter)
		if err != nil {
			fmt.Println("WEBUI ERROR", err)
		}
		return err
	})
	return wg.Wait()
}

func (m *Manager) watch(ctx context.Context, dir string, report reporter.Reporter) error {
	watcher, err := newBatcher(ctx)
	if err != nil {
		return fmt.Errorf("unable to create watcher: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("unable to watch directory: %w", err)
	}

	go func() {
		for err := range watcher.Errors {
			report.Error("watcher error: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events():
			if !ok {
				return nil
			}

			if event.Op.Has(fsnotify.Write) {
				if filepath.Base(event.Name) == specFilename {
					continue
				}

				report.RunFile(ctx, event.Name, func(r reporter.Reporter) {
					s := newSuite(m, r)
					s.run(ctx, event.Name)
				})
			} else if event.Op.Has(fsnotify.Remove) {
				if sse, ok := report.(*webui.SSEReporter); ok {
					sse.RemoveFile(event.Name)
				}
			}

		}
	}
}

func (m *Manager) GenerateSpec(dir string) error {
	path := filepath.Join(dir, specFilename)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", path, err)
	}
	defer f.Close()

	GenerateSpec(f, m.runners, m.newConfigFn(), m.helpers)
	return nil
}

func (m *Manager) doSetup(ctx context.Context, config any) (runners []spec.Runner, close func(), err error) {
	return m.setup(ctx, m.dir, config)
}

func (m *Manager) AddHelper(helper *spec.Function) {
	m.helpers = append(m.helpers, helper)
}
