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

func DoIt(ctx context.Context, conn *pgx.Conn, id int64) (pgx.CommandTag, error) {

}`, buf.String())
}
