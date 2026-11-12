// موديل جدول "EmployeePermission" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeePermission",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "permissionId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "EmployeePermission_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeePermission_employeeId_permissionId_key", Def: "UNIQUE (\"employeeId\", \"permissionId\")"},
			{Name: "EmployeePermission_permissionId_fkey", Def: "FOREIGN KEY (\"permissionId\") REFERENCES \"Permission\"(id) ON DELETE CASCADE"},
			{Name: "EmployeePermission_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
