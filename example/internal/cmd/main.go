package cmd

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nais/tester/example/internal/graph"
)

func Run(ctx context.Context) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%v/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
