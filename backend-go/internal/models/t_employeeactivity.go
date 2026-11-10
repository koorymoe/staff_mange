// موديل جدول "EmployeeActivity" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeActivity",
		Columns: []Column{
			{Name: "id", Type: "bigint", NotNull: true, Default: "nextval('\"EmployeeActivity_id_seq\"'::regclass)"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "method", Type: "text", NotNull: true},
			{Name: "pattern", Type: "text", NotNull: true},
			{Name: "path", Type: "text", NotNull: true},
			{Name: "status", Type: "integer", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeActivity_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"EmployeeActivity_emp_time_idx\" ON public.\"EmployeeActivity\" USING btree (\"employeeId\", \"createdAt\")",
			"CREATE INDEX \"EmployeeActivity_path_idx\" ON public.\"EmployeeActivity\" USING btree (path text_pattern_ops)",
		},
	})
}
