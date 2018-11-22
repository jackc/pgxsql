package generator_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx"
	"github.com/jackc/pgxsql/generator"
	"github.com/stretchr/testify/assert"
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

func TestGeneratorGenerate(t *testing.T) {

}

func TestParseSimple(t *testing.T) {
	src := `/* pgxsql
package = "main"
name = "GetSomething"
pgxsql */

select 1;
`

	qs, err := generator.Parse(src)
	assert.NoError(t, err)

	assert.Equal(t, `main`, qs.Package)
	assert.Equal(t, `GetSomething`, qs.Name)
	assert.Equal(t, `select 1;`, qs.SQL)
}
