// موديل جدول "SimCliGrammar" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimCliGrammar",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "os", Type: "text", NotNull: true},
			{Name: "tree", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "status", Type: "text", NotNull: true, Default: "'DRAFT'::text"},
			{Name: "version", Type: "integer", NotNull: true, Default: "1"},
			{Name: "sourceRef", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimCliGrammar_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
