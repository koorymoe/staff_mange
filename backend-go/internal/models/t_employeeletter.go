// موديل جدول "EmployeeLetter" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeLetter",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "addressedTo", Type: "text", NotNull: true, Default: "'السيد المدير المحترم'::text"},
			{Name: "subject", Type: "text", NotNull: true},
			{Name: "body", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "decisionNote", Type: "text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeLetter_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "EmployeeLetter_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeeLetter_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"EmployeeLetter_employee_idx\" ON public.\"EmployeeLetter\" USING btree (\"employeeId\", \"createdAt\" DESC)",
			"CREATE INDEX \"EmployeeLetter_pending_idx\" ON public.\"EmployeeLetter\" USING btree (status) WHERE (status = 'PENDING'::text)",
		},
	})
}
