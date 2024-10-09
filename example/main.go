package main

import (
	"context"

	"github.com/nais/tester/example/internal/cmd"
)

func main() {
	cmd.Run(context.Background())
}
