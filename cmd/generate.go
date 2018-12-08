package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgxsql/generator"
	"github.com/jackc/pgxsql/query"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConn() (*pgx.Conn, error) {
	config, err := pgx.ParseURI(viper.GetString("database_uri"))
	if err != nil {
		return nil, err
	}

	conn, err := pgx.Connect(config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go from SQL",
	Run: func(cmd *cobra.Command, args []string) {
		sql, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read source file failed: ", err)
			os.Exit(1)
		}

		conn, err := NewConn()
		if err != nil {
			fmt.Fprintln(os.Stderr, "connect to PostgreSQL server failed:", err)
			os.Exit(1)
		}

		query, err := query.New(string(sql), conn)
		if err != nil {
			fmt.Fprintln(os.Stderr, "parse source file failed:", err)
			os.Exit(1)
		}

		err = generator.Generate(os.Stdout, query)
		if err != nil {
			fmt.Fprintln(os.Stderr, "generate failed:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().StringP("database-uri", "d", "", "Database URI")
	viper.BindPFlag("database_uri", generateCmd.PersistentFlags().Lookup("database-uri"))
}
