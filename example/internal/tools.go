//go:build tools
// +build tools

package internal

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
)
