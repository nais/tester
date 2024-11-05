package webui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static/*
var static embed.FS

type Option func(*options)

type options struct {
	root fs.FS
}

func WithRoot(root fs.FS) Option {
	return func(o *options) {
		o.root = root
	}
}

func Run(ctx context.Context, reporter *SSEReporter, opts ...Option) error {
	o := options{
		root: static,
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.root == static {
		f, err := fs.Sub(static, "static")
		if err != nil {
			return err
		}
		o.root = f
	}

	mux := http.NewServeMux()
	mux.Handle("/", &noCache{handler: http.FileServerFS(o.root)})
	mux.Handle("/events", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle Server sent events
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := make(chan *SSEMessage, 2)
		reporter.cache.AddListener(ch)
		defer reporter.cache.RemoveListener(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				b, err := json.Marshal(msg)
				if err != nil {
					log.Println(err)
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", string(b))
				w.(http.Flusher).Flush()
			}
		}
	}))

	srv := &http.Server{
		Addr:    ":9876",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	fmt.Println("Listening on :9876")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

type noCache struct {
	handler http.Handler
}

func (n *noCache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	n.handler.ServeHTTP(w, r)
}
