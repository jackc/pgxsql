package generator_test

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgxsql/generator"
	"github.com/jackc/pgxsql/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorSimplest(t *testing.T) {
	buf := &strings.Builder{}
	q := &query.Query{
		Package:    "foo",
		Name:       "DoIt",
		ResultType: "commandTag",
		SQL:        "delete from widgets where id=$1",
		Parameters: []*query.Parameter{
			&query.Parameter{Name: "id", DataType: pgtype.Int8OID, Ordinal: 1},
		},
	}

	err := generator.Generate(buf, q)
	require.NoError(t, err)

	assert.Equal(t, `package foo

import (
  "context"

  "github.com/jackc/pgx"
  "github.com/jackc/pgxsql/base"
)

func DoIt(ctx context.Context, conn base.Execer, id int64) (pgx.CommandTag, error) {
  return conn.ExecEx(ctx, `+"`"+`delete from widgets where id=$1`+"`"+`, nil, id)
}`, buf.String())
}
