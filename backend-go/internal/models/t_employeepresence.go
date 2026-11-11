// موديل جدول "EmployeePresence" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeePresence",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "lastSeenAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "EmployeePresence_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeePresence_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
		},
	})
}
