package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	"github.com/nais/tester/example/internal/integration"
	"github.com/nais/tester/lua"
)

func main() {
	dir := filepath.Join(".", "internal", "integration", "testdata")
	ui := false
	flag.StringVar(&dir, "d", dir, "write spec to this directory")
	flag.BoolVar(&ui, "ui", ui, "enable UI")
	flag.Parse()

	mgr, err := integration.TestRunner(false)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	if ui {
		if err := mgr.RunUI(ctx, dir); err != nil {
			panic(err)
		}
		return
	}

	if err := mgr.Run(ctx, dir, lua.NewJSONReporter(os.Stdout)); err != nil {
		panic(err)
	}
}
