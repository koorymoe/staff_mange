// schemacheck — «ون بيس»: كل عمود بالقاعدة لازم يكون إله مكان بموديلات Go (قاعدة
// دازاي: «الأعمدة الجديدة تنضاف لموديلات Go»). يقرا وسوم db:"…" من كل كود
// internal/، ويقارنها بأعمدة القاعدة (information_schema).
//
// DATABASE_URL=… go run ./cmd/schemacheck [-v]
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/models"
)

// Tags كل أسماء db:"…" بالكود (أي struct)، ومعها أعمدة الموديلات (internal/models) —
// من ون بيس (11-10) الموديل هو مكان العمود الرسمي.
func Tags(root string) (map[string]bool, error) {
	out := models.ColumnNames()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(token.NewFileSet(), p, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if fld, ok := n.(*ast.Field); ok && fld.Tag != nil {
				raw, _ := strconv.Unquote(fld.Tag.Value)
				if v := reflect.StructTag(raw).Get("db"); v != "" && v != "-" {
					out[strings.Split(v, ",")[0]] = true
				}
			}
			return true
		})
		return nil
	})
	return out, err
}

type Missing struct{ Table, Column string }

// Check الأعمدة الي ماكو إلها وسم بأي موديل.
func Check(db *sqlx.DB, tags map[string]bool) ([]Missing, error) {
	rows := []struct {
		T string `db:"t"`
		C string `db:"c"`
	}{}
	if err := db.Select(&rows, `SELECT c.table_name AS t, c.column_name AS c FROM information_schema.columns c
		JOIN information_schema.tables t ON t.table_name = c.table_name AND t.table_schema = c.table_schema
		WHERE c.table_schema = 'public' AND t.table_type = 'BASE TABLE' AND c.table_name NOT LIKE 'schema_%' AND c.table_name <> 'SchemaMigration'`); err != nil {
		return nil, err
	}
	out := []Missing{}
	for _, r := range rows {
		if !tags[r.C] {
			out = append(out, Missing{r.T, r.C})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Table+out[i].Column < out[j].Table+out[j].Column })
	return out, nil
}

func main() {
	tags, err := Tags("internal")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	db := sqlx.MustConnect("postgres", os.Getenv("DATABASE_URL"))
	miss, err := Check(db, tags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("وسوم db بالكود: %d · أعمدة بلا موديل: %d\n", len(tags), len(miss))
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		for _, m := range miss {
			fmt.Printf("  %s.%s\n", m.Table, m.Column)
		}
	}
}
