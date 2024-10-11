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
	flag.StringVar(&dir, "d", dir, "write spec to this directory")
	flag.Parse()

	mgr, err := integration.TestRunner(false)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	if err := mgr.Run(ctx, dir, lua.NewJSONReporter(os.Stdout)); err != nil {
		panic(err)
	}
}
