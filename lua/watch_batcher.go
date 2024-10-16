package lua

import (
	"context"
	"time"

	"github.com/fsnotify/fsnotify"
)

type batcher struct {
	*fsnotify.Watcher
	changed map[string]fsnotify.Event
	events  chan fsnotify.Event
}

func newBatcher(ctx context.Context) (*batcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	b := &batcher{
		Watcher: watcher,
		changed: make(map[string]fsnotify.Event),
		events:  make(chan fsnotify.Event),
	}

	go b.run(ctx)

	return b, nil
}

func (b *batcher) run(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-b.Watcher.Events:
			switch {
			case event.Has(fsnotify.Write) || event.Has(fsnotify.Create):
				b.changed[event.Name] = event
			case event.Has(fsnotify.Remove):
				delete(b.changed, event.Name)
			case event.Has(fsnotify.Rename):
				delete(b.changed, event.Name)
			}
		case <-ticker.C:
			for _, event := range b.changed {
				b.events <- event
			}
			b.changed = make(map[string]fsnotify.Event)
		}
	}
}

func (b *batcher) Events() <-chan fsnotify.Event {
	return b.events
}

func (b *batcher) Close() error {
	close(b.events)
	return b.Watcher.Close()
}
