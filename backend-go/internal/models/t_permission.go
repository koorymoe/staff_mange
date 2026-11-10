// موديل جدول "Permission" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Permission",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "label", Type: "text", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "Permission_name_key", Def: "UNIQUE (name)"},
			{Name: "Permission_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
