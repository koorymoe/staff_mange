// موديل جدول "EmployeeNegligence" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeNegligence",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "day", Type: "date", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "detail", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "reason", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "compensatedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "voidedAt", Type: "timestamp with time zone"},
			{Name: "voidNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeNegligence_employeeId_day_kind_key", Def: "UNIQUE (\"employeeId\", day, kind)"},
			{Name: "EmployeeNegligence_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeeNegligence_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"EmployeeNegligence_day\" ON public.\"EmployeeNegligence\" USING btree (day DESC)",
		},
	})
}
