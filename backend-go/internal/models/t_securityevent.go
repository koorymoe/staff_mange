// موديل جدول "SecurityEvent" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SecurityEvent",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text"},
			{Name: "employeeName", Type: "text"},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "detail", Type: "text"},
			{Name: "ip", Type: "text"},
			{Name: "userAgent", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SecurityEvent_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SecurityEvent_createdAt_idx\" ON public.\"SecurityEvent\" USING btree (\"createdAt\" DESC)",
			"CREATE INDEX \"SecurityEvent_employee_idx\" ON public.\"SecurityEvent\" USING btree (\"employeeId\")",
		},
	})
}
