package database

import (
	"context"
	"database/sql"
	"time"
)

type delayingDBTX struct {
	inner DBTX
}

func (d *delayingDBTX) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	time.Sleep(5 * time.Second)
	return d.inner.ExecContext(ctx, query, args...)
}

func (d *delayingDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	time.Sleep(5 * time.Second)
	return d.inner.PrepareContext(ctx, query)
}

func (d *delayingDBTX) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	time.Sleep(5 * time.Second)
	return d.inner.QueryContext(ctx, query, args...)
}

func (d *delayingDBTX) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	time.Sleep(5 * time.Second)
	return d.inner.QueryRowContext(ctx, query, args...)
}

func NewDelayedDB(db DBTX) *Queries {
	return &Queries{db: &delayingDBTX{inner: db}}
}
