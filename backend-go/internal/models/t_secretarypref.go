// موديل جدول "SecretaryPref" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SecretaryPref",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SecretaryPref_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "SecretaryPref_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
