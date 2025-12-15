package webui

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/nais/tester/lua/reporter"
)

type TestError struct {
	Message string `json:"message"`
}

type Test struct {
	Filename string `json:"filename"`
	Name     string `json:"name"`
	Runner   string `json:"runner"`
	lock     sync.RWMutex
	Errors   []*TestError  `json:"errors"`
	Duration time.Duration `json:"duration"`

	start time.Time
	cache *sseCache
}

func (t *Test) Start() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.Errors = nil
	t.start = time.Now()

	t.cache.Broadcast(&SSEMessage{
		Type: "start_test",
		Data: t,
	})
}

func (t *Test) End() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.Duration = time.Since(t.start)

	t.cache.Broadcast(&SSEMessage{
		Type: "end_test",
		Data: t,
	})
}

func (t *Test) AddError(msg string, args ...any) {
	if t == nil {
		fmt.Printf(msg, args...)
		fmt.Println()
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()

	t.Errors = append(t.Errors, &TestError{
		Message: msg,
	})

	t.cache.Broadcast(&SSEMessage{
		Type: "error",
		Data: t,
	})
}

type File struct {
	Name     string `json:"name"`
	lock     sync.RWMutex
	SubTests []*Test       `json:"subTests"`
	Duration time.Duration `json:"duration"`

	start time.Time
	cache *sseCache
}

func (f *File) Start() {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.SubTests = nil
	f.start = time.Now()

	f.cache.Broadcast(&SSEMessage{
		Type: "start",
		Data: f,
	})
}

func (f *File) End() {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.Duration = time.Since(f.start)

	f.cache.Broadcast(&SSEMessage{
		Type: "end",
		Data: f,
	})
}

func (f *File) AddTest(name, runner string) *Test {
	f.lock.Lock()
	defer f.lock.Unlock()

	test := &Test{
		Filename: f.Name,
		Name:     name,
		Runner:   runner,
		cache:    f.cache,
	}

	f.SubTests = append(f.SubTests, test)
	return test
}

type SSEMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type listener chan *SSEMessage

type sseCache struct {
	dirPrefix string

	lock  sync.RWMutex
	files map[string]*File

	listeners []listener
}

func (c *sseCache) Broadcast(msg *SSEMessage) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, ch := range c.listeners {
		ch <- msg
	}
}

func (c *sseCache) RemoveFile(name string) {
	c.lock.Lock()
	delete(c.files, name)
	c.lock.Unlock()

	c.Broadcast(&SSEMessage{
		Type: "remove",
		Data: name,
	})
}

func (c *sseCache) AddListener(ch listener) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ch <- &SSEMessage{
		Type: "init",
		Data: c.files,
	}
	c.listeners = append(c.listeners, ch)
}

func (c *sseCache) RemoveListener(ch listener) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.listeners = slices.DeleteFunc(c.listeners, func(e listener) bool {
		return e == ch
	})
}

func (c *sseCache) AddFile(name string) *File {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.files == nil {
		c.files = make(map[string]*File)
	}

	if _, ok := c.files[name]; ok {
		return c.files[name]
	}

	rel, _ := filepath.Rel(c.dirPrefix, name)

	file := &File{
		Name:  rel,
		cache: c,
	}

	c.files[name] = file
	return file
}

type SSEReporter struct {
	cache *sseCache
	file  *File
	test  *Test
}

func NewSSEReporter(dir string) *SSEReporter {
	return &SSEReporter{
		cache: &sseCache{
			dirPrefix: dir,
		},
	}
}

func (r *SSEReporter) RunFile(ctx context.Context, filename string, fn func(reporter.Reporter)) {
	file := r.cache.AddFile(filename)
	file.Start()
	fn(&SSEReporter{file: file})
	file.End()
}

func (r *SSEReporter) RunTest(ctx context.Context, runner, name string, fn func(reporter.Reporter)) {
	test := r.file.AddTest(name, runner)
	test.Start()
	fn(&SSEReporter{file: r.file, test: test})
	test.End()
}

func (r *SSEReporter) Error(msg string, args ...any) {
	r.test.AddError(msg, args...)
}

func (r *SSEReporter) RemoveFile(name string) {
	r.cache.RemoveFile(name)
}
