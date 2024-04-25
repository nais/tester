package graph

import "github.com/nais/tester/example/internal/database"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db *database.Queries
}

func NewResolver(db *database.Queries) *Resolver {
	return &Resolver{db: db}
}
