// موديل جدول "StaffRequestEmployee" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "StaffRequestEmployee",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "requestId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "StaffRequestEmployee_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "StaffRequestEmployee_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "StaffRequestEmployee_requestId_employeeId_key", Def: "UNIQUE (\"requestId\", \"employeeId\")"},
			{Name: "StaffRequestEmployee_requestId_fkey", Def: "FOREIGN KEY (\"requestId\") REFERENCES \"StaffRequest\"(id) ON DELETE CASCADE"},
		},
	})
}
