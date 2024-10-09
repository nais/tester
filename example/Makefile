.PHONY: generate
generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
	go run github.com/99designs/gqlgen generate
	go run ./internal/tools/tester_spec
