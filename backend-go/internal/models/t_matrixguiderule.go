// موديل جدول "MatrixGuideRule" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixGuideRule",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "route", Type: "text", NotNull: true},
			{Name: "match", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "groups", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "onlyIfPending", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "priority", Type: "integer", NotNull: true, Default: "0"},
			{Name: "enabled", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "hits", Type: "integer", NotNull: true, Default: "0"},
			{Name: "lastHitAt", Type: "timestamp with time zone"},
			{Name: "source", Type: "text", NotNull: true, Default: "'ADMIN'::text"},
		},
		Constraints: []Constraint{
			{Name: "MatrixGuideRule_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
