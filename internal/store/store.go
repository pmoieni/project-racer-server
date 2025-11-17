package store

import (
	"context"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var (
	ErrConflict = errors.New("store: entity already exists")
	ErrNotFound = errors.New("store: entity not found")
)

type StoreErr struct {
	Code    int
	Message string
	Err     error
}

func (e *StoreErr) Error() string {
	return fmt.Sprintf(
		"[Store Error]\nError Code: %d\nError Message: %s\nError: %s\n",
		e.Code,
		e.Message,
		e.Err,
	)
}

func NewDB(ctx context.Context, dsn string) (*sqlx.DB, error) {
	// TODO: pass context
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	pgxdb := stdlib.OpenDBFromPool(pool)
	return sqlx.NewDb(pgxdb, "pgx"), nil
}
