package generator

import (
	"io"
	"strings"
	"text/template"

	"github.com/jackc/pgxsql/query"
)

// Generate produces Go code from q and writes it to w.
//
// q is mutated / consumed by this operation and must not be used again.
func Generate(w io.Writer, q *query.Query) error {
	q.SQL = "`" + strings.Replace(q.SQL, "`", "` + `", -1) + "`"

	const src = `package {{.Package}}

import (
  "context"

  "github.com/jackc/pgx"
  "github.com/jackc/pgxsql/base"
)

func {{.Name}}(ctx context.Context, conn base.Execer, id int64) (pgx.CommandTag, error) {
  return conn.ExecEx(ctx, {{.SQL}}, nil, id)
}`
	t := template.Must(template.New("src").Parse(src))

	return t.Execute(w, q)
}
