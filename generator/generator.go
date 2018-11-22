package generator

import (
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx"
)

type Generator struct {
	conn *pgx.Conn
}

func New(conn *pgx.Conn) *Generator {
	return &Generator{conn: conn}
}

type QuerySource struct {
	Package string
	Name    string
	SQL     string
}

func Parse(src string) (*QuerySource, error) {
	configToml, sql := splitConfigAndSQL(src)
	qs := &QuerySource{SQL: sql}
	_, err := toml.Decode(configToml, &qs)

	return qs, err
}

func splitConfigAndSQL(src string) (config, sql string) {
	re := regexp.MustCompile(`(?s:/\* pgxsql.+pgxsql \*/)`)
	sql = re.ReplaceAllStringFunc(src, func(match string) string {
		config += match[9 : len(match)-9]
		return ""
	})

	return strings.TrimSpace(config), strings.TrimSpace(sql)
}
