package base

import (
	"context"

	"github.com/jackc/pgx"
)

type Execer interface {
	ExecEx(ctx context.Context, sql string, options *pgx.QueryExOptions, arguments ...interface{}) (pgx.CommandTag, error)
}
