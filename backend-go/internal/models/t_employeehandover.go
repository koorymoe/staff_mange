// موديل جدول "EmployeeHandover" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeHandover",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "fromId", Type: "text", NotNull: true},
			{Name: "toId", Type: "text", NotNull: true},
			{Name: "byId", Type: "text"},
			{Name: "summary", Type: "jsonb", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeHandover_byId_fkey", Def: "FOREIGN KEY (\"byId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "EmployeeHandover_fromId_fkey", Def: "FOREIGN KEY (\"fromId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeeHandover_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "EmployeeHandover_toId_fkey", Def: "FOREIGN KEY (\"toId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
		},
	})
}
