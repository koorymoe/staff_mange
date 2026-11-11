// modelgen — 🏴‍☠️ ون بيس: يولّد موديلات internal/models من القاعدة الحالية (مرة وحدة
// بالبداية، أو لمن تتبنّى جدول انسوى برّا الموديلات). بعدها الموديل هو المصدر.
//
// DATABASE_URL=… go run ./cmd/modelgen
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/models"
)

var nonWord = regexp.MustCompile(`[^A-Za-z0-9]+`)

func fileName(t string) string {
	return "t_" + strings.ToLower(nonWord.ReplaceAllString(t, "_")) + ".go"
}

func main() {
	db := sqlx.MustConnect("postgres", os.Getenv("DATABASE_URL"))
	if err := database.Migrate(db, "", ""); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	tabs, enums, err := models.Introspect(db)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dir := "internal/models"
	old, _ := filepath.Glob(filepath.Join(dir, "t_*.go"))
	for _, f := range old {
		_ = os.Remove(f)
	}
	q := strconv.Quote
	for _, t := range tabs {
		var b strings.Builder
		fmt.Fprintf(&b, "// موديل جدول %s — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.\n\npackage models\n\n", q(t.Name))
		fmt.Fprintf(&b, "func init() {\n\tregister(Table{\n\t\tName: %s,\n\t\tColumns: []Column{\n", q(t.Name))
		for _, c := range t.Columns {
			fmt.Fprintf(&b, "\t\t\t{Name: %s, Type: %s", q(c.Name), q(c.Type))
			if c.NotNull {
				b.WriteString(", NotNull: true")
			}
			if c.Default != "" {
				fmt.Fprintf(&b, ", Default: %s", q(c.Default))
			}
			b.WriteString("},\n")
		}
		b.WriteString("\t\t},\n")
		if len(t.Constraints) > 0 {
			b.WriteString("\t\tConstraints: []Constraint{\n")
			for _, c := range t.Constraints {
				fmt.Fprintf(&b, "\t\t\t{Name: %s, Def: %s},\n", q(c.Name), q(c.Def))
			}
			b.WriteString("\t\t},\n")
		}
		if len(t.Indexes) > 0 {
			b.WriteString("\t\tIndexes: []string{\n")
			for _, x := range t.Indexes {
				fmt.Fprintf(&b, "\t\t\t%s,\n", q(x))
			}
			b.WriteString("\t\t},\n")
		}
		b.WriteString("\t})\n}\n")
		if err := os.WriteFile(filepath.Join(dir, fileName(t.Name)), []byte(b.String()), 0o644); err != nil {
			panic(err)
		}
	}
	var b strings.Builder
	b.WriteString("// الأنواع الخاصة (ENUM) — نفس قاعدة الموديلات: القيمة الجديدة تنضاف هنا.\n\npackage models\n\nfunc init() {\n")
	for _, e := range enums {
		vals := make([]string, len(e.Values))
		for i, v := range e.Values {
			vals[i] = q(v)
		}
		fmt.Fprintf(&b, "\tregisterEnum(Enum{Name: %s, Values: []string{%s}})\n", q(e.Name), strings.Join(vals, ", "))
	}
	b.WriteString("}\n")
	if err := os.WriteFile(filepath.Join(dir, "t__enums.go"), []byte(b.String()), 0o644); err != nil {
		panic(err)
	}
	objs, err := models.IntrospectObjects(db)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var o strings.Builder
	o.WriteString("// الإضافات والتسلسلات والدوال والـviews والـtriggers — جزء من الموديلات (القاعدة تنبني منها كاملة).\n\npackage models\n\nfunc init() {\n")
	for _, x := range objs {
		fmt.Fprintf(&o, "\tregisterObject(Object{Kind: %s, Key: %s, Def: %s})\n", q(x.Kind), q(x.Key), q(x.Def))
	}
	o.WriteString("}\n")
	if err := os.WriteFile(filepath.Join(dir, "t__objects.go"), []byte(o.String()), 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("انولد %d موديل جدول و%d نوع و%d شي ثاني\n", len(tabs), len(enums), len(objs))
}
