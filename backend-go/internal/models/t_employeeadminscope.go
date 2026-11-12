// موديل جدول "EmployeeAdminScope" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeAdminScope",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "roles", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeAdminScope_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeeAdminScope_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
			{Name: "EmployeeAdminScope_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
