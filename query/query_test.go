package query_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgxsql/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var conn *pgx.Conn

func TestMain(m *testing.M) {
	databaseURI := os.Getenv("TEST_DATABASE_URI")
	if databaseURI == "" {
		fmt.Fprintln(os.Stderr, "TEST_DATABASE_URI environment variable is required but not set")
		os.Exit(1)
	}

	connConfig, err := pgx.ParseURI(databaseURI)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to parse TEST_DATABASE_URI:", err)
		os.Exit(1)
	}

	conn, err = pgx.Connect(connConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to connect to PostgreSQL server:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestNewSimplest(t *testing.T) {
	src := `/* pgxsql
package = "main"
name = "GetSomething"
pgxsql */

select 1;
`

	qs, err := query.New(src, conn)
	assert.NoError(t, err)

	assert.Equal(t, `main`, qs.Package)
	assert.Equal(t, `GetSomething`, qs.Name)
	assert.Equal(t, `select 1;`, qs.SQL)
	assert.Empty(t, qs.Parameters)
}

func TestNewQueryParameter(t *testing.T) {
	src := `/* pgxsql
package = "main"
name = "GetPerson"
pgxsql */

select name
from person
where id = :id;
`

	qs, err := query.New(src, conn)
	require.NoError(t, err)
	require.NotNil(t, qs)

	assert.Equal(t, `main`, qs.Package)
	assert.Equal(t, `GetPerson`, qs.Name)
	assert.Equal(t, `select name
from person
where id = $1;`, qs.SQL)
	require.Len(t, qs.Parameters, 1)
	assert.Equal(t, "id", qs.Parameters[0].Name)
	assert.Equal(t, pgtype.OID(pgtype.Int8OID), qs.Parameters[0].DataType)
	assert.Equal(t, uint16(1), qs.Parameters[0].Ordinal)
}

func TestNewQuotedAlmostQueryParameters(t *testing.T) {
	src := `/* pgxsql
package = "main"
name = "GetSomething"
pgxsql */

select ':a' as ":b";
`

	qs, err := query.New(src, conn)
	assert.NoError(t, err)

	assert.Equal(t, `main`, qs.Package)
	assert.Equal(t, `GetSomething`, qs.Name)
	assert.Equal(t, `select ':a' as ":b";`, qs.SQL)
	assert.Empty(t, qs.Parameters)
}
