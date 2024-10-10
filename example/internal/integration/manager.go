package integration

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	// allow test containers to trigger postgres snapshots
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nais/tester/example/internal/database"
	"github.com/nais/tester/example/internal/graph"
	testmanager "github.com/nais/tester/lua"
	"github.com/nais/tester/lua/runner"
	"github.com/nais/tester/lua/spec"
	"github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRunner() (*testmanager.Manager, error) {
	ctx := context.Background()
	mgr, err := testmanager.New(newConfig, newManager(ctx), &runner.GQL{}, &runner.SQL{}, &runner.REST{})
	if err != nil {
		return nil, err
	}

	// if err := mgr.Run(ctx, os.DirFS("./testdata")); err != nil {
	// 	return nil, err
	// }

	return mgr, nil
}

func newManager(ctx context.Context) testmanager.SetupFunc {
	container, connStr, err := startPostgresql(ctx)
	if err != nil {
		panic(err)
	}

	return func(ctx context.Context, _ string, _ any) ([]spec.Runner, func(), error) {
		ctx, done := context.WithCancel(ctx)
		cleanups := []func(){}

		db, pool, cleanup, err := newDB(ctx, container, connStr)
		if err != nil {
			done()
			return nil, nil, err
		}
		cleanups = append(cleanups, cleanup)

		log := logrus.New()

		log.Out = io.Discard
		if testing.Verbose() {
			log.Out = os.Stdout
			log.Level = logrus.DebugLevel
		}

		runners := []spec.Runner{
			newRestRunner(),
			newGQLRunner(ctx, db),
			runner.NewSQLRunner(pool),
		}

		return runners, func() {
			for _, cleanup := range cleanups {
				cleanup()
			}
			done()
		}, nil
	}
}

func newRestRunner() spec.Runner {
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"message": "hello world"}`)
	})

	return runner.NewRestRunner(router)
}

func newGQLRunner(_ context.Context, db *database.Queries) spec.Runner {
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

func startPostgresql(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	container, err := postgres.Run(ctx, "docker.io/postgres:16-alpine",
		// testcontainers.WithLogger(testcontainers.TestLogger(t)),
		postgres.WithDatabase("example"),
		postgres.WithUsername("example"),
		postgres.WithPassword("example"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2)),
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start container: %w", err)
	}

	// t.Cleanup(func() {
	// 	if err := container.Terminate(ctx); err != nil {
	// 		log.Fatalf("failed to terminate container: %s", err)
	// 	}
	// })

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

	pool.Reset()

	if err := container.Snapshot(ctx, postgres.WithSnapshotName("migrated")); err != nil {
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
