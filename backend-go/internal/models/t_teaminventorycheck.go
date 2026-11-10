// موديل جدول "TeamInventoryCheck" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TeamInventoryCheck",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "leaderId", Type: "text", NotNull: true},
			{Name: "employee1Id", Type: "text"},
			{Name: "employee2Id", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "TeamInventoryCheck_employee1Id_fkey", Def: "FOREIGN KEY (\"employee1Id\") REFERENCES \"Employee\"(id)"},
			{Name: "TeamInventoryCheck_employee2Id_fkey", Def: "FOREIGN KEY (\"employee2Id\") REFERENCES \"Employee\"(id)"},
			{Name: "TeamInventoryCheck_leaderId_fkey", Def: "FOREIGN KEY (\"leaderId\") REFERENCES \"Employee\"(id)"},
			{Name: "TeamInventoryCheck_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
