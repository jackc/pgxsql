package generator

import (
	"io"
	"text/template"

	"github.com/jackc/pgxsql/query"
)

func Generate(w io.Writer, q *query.Query) error {
	const src = `package {{.Package}}

func {{.Name}}(ctx context.Context, conn *pgx.Conn, id int64) (pgx.CommandTag, error) {

}`
	t := template.Must(template.New("src").Parse(src))

	return t.Execute(w, q)
}
