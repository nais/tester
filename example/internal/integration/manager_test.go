package integration_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nais/tester/example/internal/database"
	"github.com/nais/tester/example/internal/graph"
	"github.com/nais/tester/testmanager"
	"github.com/nais/tester/testmanager/runner"
	"github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRunner(t *testing.T) {
	ctx := context.Background()
	mgr := testmanager.New(t, newManager(ctx, t))

	if err := mgr.Run(ctx, os.DirFS("./testdata")); err != nil {
		t.Fatal(err)
	}
}

func newManager(ctx context.Context, t *testing.T) testmanager.CreateRunnerFunc[*Config] {
	container, connStr, err := startPostgresql(ctx, t)
	if err != nil {
		t.Fatal(err)
	}

	return func(ctx context.Context, config *Config, state map[string]any) ([]testmanager.Runner, func(), []testmanager.Option, error) {
		ctx, done := context.WithCancel(ctx)
		cleanups := []func(){}

		opts := []testmanager.Option{}

		db, pool, cleanup, err := newDB(ctx, container, connStr)
		if err != nil {
			done()
			return nil, nil, opts, err
		}
		cleanups = append(cleanups, cleanup)

		log := logrus.New()

		log.Out = io.Discard
		if testing.Verbose() {
			log.Out = os.Stdout
			log.Level = logrus.DebugLevel
		}

		runners := []testmanager.Runner{
			newRestRunner(ctx, t),
			newGQLRunner(ctx, t, db),
			runner.NewSQLRunner(pool),
		}

		return runners, func() {
			for _, cleanup := range cleanups {
				cleanup()
			}
			done()
		}, opts, nil
	}
}

func newRestRunner(ctx context.Context, t *testing.T) testmanager.Runner {
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"message": "hello world"}`)
	})

	return runner.NewRestRunner(router)
}

func newGQLRunner(_ context.Context, _ *testing.T, db *database.Queries) testmanager.Runner {
	log := logrus.New()
	log.Out = io.Discard

	resolver := graph.NewResolver(db)

	newServer := func(es graphql.ExecutableSchema) *handler.Server {
		srv := handler.New(es)
		srv.AddTransport(transport.SSE{})
		srv.AddTransport(transport.GET{})
		srv.AddTransport(transport.POST{})

		return srv
	}

	srv := newServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	return runner.NewGQLRunner(srv)
}

func startPostgresql(ctx context.Context, t *testing.T) (*postgres.PostgresContainer, string, error) {
	container, err := postgres.RunContainer(ctx,
		testcontainers.WithLogger(testcontainers.TestLogger(t)),
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase("example"),
		postgres.WithUsername("example"),
		postgres.WithPassword("example"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2)),
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start container: %w", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create pool: %w", err)
	}

	if err := database.Migrate(ctx, pool); err != nil {
		return nil, "", fmt.Errorf("failed to migrate: %w", err)
	}

	if err := container.Snapshot(ctx); err != nil {
		return nil, "", fmt.Errorf("failed to snapshot: %w", err)
	}

	return container, connStr, nil
}

func newDB(ctx context.Context, container *postgres.PostgresContainer, connStr string) (*database.Queries, *pgxpool.Pool, func(), error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create pool: %w", err)
	}

	queries := database.New(pool)

	cleanup := func() {
		pool.Close()
		if err := container.Restore(ctx); err != nil {
			log.Fatalf("failed to restore: %s", err)
		}
	}

	return queries, pool, cleanup, nil
}
