// موديل جدول "DepartmentHead" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DepartmentHead",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "departmentId", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "phone", Type: "text"},
			{Name: "employeeId", Type: "text"},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DepartmentHead_departmentId_fkey", Def: "FOREIGN KEY (\"departmentId\") REFERENCES \"Department\"(id) ON DELETE CASCADE"},
			{Name: "DepartmentHead_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "DepartmentHead_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"DepartmentHead_dept_idx\" ON public.\"DepartmentHead\" USING btree (\"departmentId\")",
		},
	})
}
