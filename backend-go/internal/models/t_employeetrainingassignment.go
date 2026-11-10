// موديل جدول "EmployeeTrainingAssignment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeTrainingAssignment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeTrainingAssignment_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeeTrainingAssignment_employeeId_serviceId_key", Def: "UNIQUE (\"employeeId\", \"serviceId\")"},
			{Name: "EmployeeTrainingAssignment_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "EmployeeTrainingAssignment_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
		},
	})
}
