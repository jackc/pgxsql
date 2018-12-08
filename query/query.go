package query

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
)

type Parameter struct {
	Name     string
	DataType pgtype.OID
	Ordinal  uint16
}

type Query struct {
	Package    string
	Name       string
	SQL        string
	Parameters []*Parameter
}

func New(src string, conn *pgx.Conn) (*Query, error) {
	configToml, sql := extractConfigAndSQL(src)
	var q Query
	_, err := toml.Decode(configToml, &q)
	if err != nil {
		return nil, err
	}

	q.SQL, q.Parameters = extractNamedParameters(sql)

	ps, err := conn.Prepare("pgxsql", q.SQL)
	if err != nil {
		return nil, err
	}
	defer conn.Deallocate("pgxsql")

	for i, oid := range ps.ParameterOIDs {
		parameterOrdinal := uint16(i + 1)

		var parameter *Parameter
		for _, p := range q.Parameters {
			if parameterOrdinal == p.Ordinal {
				parameter = p
				break
			}
		}

		if parameter == nil {
			return nil, errors.New("unable to match named parameter with prepared statement parameter OID")
		}

		parameter.DataType = oid
	}

	return &q, nil
}

func extractConfigAndSQL(src string) (config, sql string) {
	re := regexp.MustCompile(`(?s:/\* pgxsql.+pgxsql \*/)`)
	sql = re.ReplaceAllStringFunc(src, func(match string) string {
		config += match[9 : len(match)-9]
		return ""
	})

	return strings.TrimSpace(config), strings.TrimSpace(sql)
}

func extractNamedParameters(inSQL string) (string, []*Parameter) {
	l := &sqlLexer{
		src:     inSQL,
		stateFn: rawState,
	}

	for l.stateFn != nil {
		l.stateFn = l.stateFn(l)
	}

	outSQL := &strings.Builder{}

	nameToParameterMap := make(map[namedPlaceholder]*Parameter)

	for _, p := range l.parts {
		switch p := p.(type) {
		case sqlSnippet:
			outSQL.WriteString(string(p))
		case namedPlaceholder:
			var parameter *Parameter
			var present bool
			if parameter, present = nameToParameterMap[p]; !present {
				parameter = &Parameter{Name: string(p), Ordinal: uint16(len(nameToParameterMap) + 1)}
				nameToParameterMap[p] = parameter
			}
			outSQL.WriteString("$")
			outSQL.WriteString(strconv.FormatInt(int64(parameter.Ordinal), 10))
		default:
			panic("unknown part type")
		}

	}

	parameters := make([]*Parameter, 0, len(nameToParameterMap))
	for _, p := range nameToParameterMap {
		parameters = append(parameters, p)
	}

	return outSQL.String(), parameters
}

type sqlSnippet string
type namedPlaceholder string

// part is either a sqlSnippet or namedPlaceholder.
type part interface{}

type sqlLexer struct {
	src     string
	start   int
	pos     int
	stateFn stateFn
	parts   []part
}

func NewQuery(sql string) ([]part, error) {
	l := &sqlLexer{
		src:     sql,
		stateFn: rawState,
	}

	for l.stateFn != nil {
		l.stateFn = l.stateFn(l)
	}

	return l.parts, nil
}

type stateFn func(*sqlLexer) stateFn

func rawState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case 'e', 'E':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune == '\'' {
				l.pos += width
				return escapeStringState
			}
		case '\'':
			return singleQuoteState
		case '"':
			return doubleQuoteState
		case ':':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune == ':' { // type conversion
				l.pos += width
			} else { // named placeholder
				if l.pos-l.start > 0 {
					l.parts = append(l.parts, sqlSnippet(l.src[l.start:l.pos-width]))
				}
				l.start = l.pos
				return namedPlaceholderState
			}
		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, sqlSnippet(l.src[l.start:l.pos]))
				l.start = l.pos
			}
			return nil
		}
	}
}

func singleQuoteState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case '\'':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune != '\'' {
				return rawState
			}
			l.pos += width
		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, sqlSnippet(l.src[l.start:l.pos]))
				l.start = l.pos
			}
			return nil
		}
	}
}

func doubleQuoteState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case '"':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune != '"' {
				return rawState
			}
			l.pos += width
		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, sqlSnippet(l.src[l.start:l.pos]))
				l.start = l.pos
			}
			return nil
		}
	}
}

// namedPlaceholderState consumes a placeholder value. The : must have already has
// already been consumed.
func namedPlaceholderState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])

		if isNamedPlaceholderRune(r) {
			l.pos += width
		} else {
			l.parts = append(l.parts, namedPlaceholder(l.src[l.start:l.pos]))
			l.start = l.pos
			return rawState
		}
	}
}

func escapeStringState(l *sqlLexer) stateFn {
	for {
		r, width := utf8.DecodeRuneInString(l.src[l.pos:])
		l.pos += width

		switch r {
		case '\\':
			_, width = utf8.DecodeRuneInString(l.src[l.pos:])
			l.pos += width
		case '\'':
			nextRune, width := utf8.DecodeRuneInString(l.src[l.pos:])
			if nextRune != '\'' {
				return rawState
			}
			l.pos += width
		case utf8.RuneError:
			if l.pos-l.start > 0 {
				l.parts = append(l.parts, l.src[l.start:l.pos])
				l.start = l.pos
			}
			return nil
		}
	}
}

func isNamedPlaceholderRune(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || '0' <= ch && ch <= '9'
}
