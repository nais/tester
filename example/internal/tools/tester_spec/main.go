package main

import (
	"flag"
	"path/filepath"

	"github.com/nais/tester/example/internal/integration"
)

func main() {
	dir := filepath.Join(".", "internal", "integration", "testdata")
	flag.StringVar(&dir, "d", dir, "write spec to this directory")
	flag.Parse()

	mgr, err := integration.TestRunner(true)
	if err != nil {
		panic(err)
	}

	if err := mgr.GenerateSpec(dir); err != nil {
		panic(err)
	}
}
