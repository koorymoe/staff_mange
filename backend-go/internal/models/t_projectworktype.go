// موديل جدول "ProjectWorkType" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectWorkType",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ProjectWorkType_name_key", Def: "UNIQUE (name)"},
			{Name: "ProjectWorkType_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
