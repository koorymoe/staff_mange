// موديل جدول "TeamInventoryToolCatalog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TeamInventoryToolCatalog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "TeamInventoryToolCatalog_name_key", Def: "UNIQUE (name)"},
			{Name: "TeamInventoryToolCatalog_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
