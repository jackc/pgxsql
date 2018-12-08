package integrationtest_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgxsql/integrationtest"

	"github.com/jackc/pgx"
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

func TestDeleteRow(t *testing.T) {
	tx, err := conn.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	commandTag, err := integrationtest.DeletePerson(context.Background(), tx, 1)
	assert.NoError(t, err)
	assert.Equal(t, "DELETE 1", string(commandTag))

	rowFound := false
	err = conn.QueryRow("select true from person where id=$1", 1).Scan(&rowFound)
	assert.Equal(t, pgx.ErrNoRows, err)
	assert.False(t, rowFound)
}
