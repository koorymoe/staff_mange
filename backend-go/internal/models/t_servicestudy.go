// موديل جدول "ServiceStudy" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ServiceStudy",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "createdById", Type: "text", NotNull: true},
			{Name: "archived", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ServiceStudy_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "ServiceStudy_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
