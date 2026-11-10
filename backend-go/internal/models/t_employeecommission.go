// موديل جدول "EmployeeCommission" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeCommission",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "leaderInvoiceId", Type: "text", NotNull: true},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "executionCommission", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "salesCommission", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "totalCommission", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeCommission_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "EmployeeCommission_leaderInvoiceId_fkey", Def: "FOREIGN KEY (\"leaderInvoiceId\") REFERENCES \"LeaderInvoice\"(id) ON DELETE CASCADE"},
			{Name: "EmployeeCommission_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "EmployeeCommission_role_check", Def: "CHECK ((role = ANY (ARRAY['LEADER'::text, 'TECHNICIAN'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"EmployeeCommission_createdAt_idx\" ON public.\"EmployeeCommission\" USING btree (\"createdAt\")",
			"CREATE INDEX \"EmployeeCommission_employeeId_idx\" ON public.\"EmployeeCommission\" USING btree (\"employeeId\")",
			"CREATE INDEX \"EmployeeCommission_leaderInvoiceId_idx\" ON public.\"EmployeeCommission\" USING btree (\"leaderInvoiceId\")",
		},
	})
}
