// موديل جدول "DesignTool" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DesignTool",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "discipline", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "url", Type: "text", NotNull: true},
			{Name: "howTo", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "sort", Type: "integer", NotNull: true, Default: "0"},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
		},
		Constraints: []Constraint{
			{Name: "DesignTool_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
