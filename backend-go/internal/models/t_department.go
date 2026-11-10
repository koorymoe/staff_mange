// موديل جدول "Department" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Department",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "Department_name_key", Def: "UNIQUE (name)"},
			{Name: "Department_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
