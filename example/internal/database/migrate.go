package database

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/schema.sql
var scheme []byte

func Migrate(ctx context.Context, conn *pgxpool.Pool) error {
	_, err := conn.Exec(ctx, string(scheme))
	return err
}
