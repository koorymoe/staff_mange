// موديل جدول "DisciplinePoints" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DisciplinePoints",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "points", Type: "integer", NotNull: true, Default: "100"},
			{Name: "lastRestoredAt", Type: "timestamp with time zone"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DisciplinePoints_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "DisciplinePoints_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
		},
	})
}
