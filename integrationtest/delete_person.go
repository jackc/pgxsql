package integrationtest

import (
  "context"

  "github.com/jackc/pgx"
  "github.com/jackc/pgxsql/base"
)

func DeletePerson(ctx context.Context, conn base.Execer, id int64) (pgx.CommandTag, error) {
  return conn.ExecEx(ctx, `delete from person where id=$1;`, nil, id)
}