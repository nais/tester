.PHONY: generate
generate:
	go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate
	go tool github.com/99designs/gqlgen generate
	go run ./internal/tools/tester_spec
