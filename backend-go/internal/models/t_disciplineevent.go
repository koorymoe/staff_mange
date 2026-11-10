// موديل جدول "DisciplineEvent" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DisciplineEvent",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "delta", Type: "integer", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "byEmployeeId", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "DisciplineEvent_byEmployeeId_fkey", Def: "FOREIGN KEY (\"byEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "DisciplineEvent_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "DisciplineEvent_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"DisciplineEvent_byEmployee_idx\" ON public.\"DisciplineEvent\" USING btree (\"byEmployeeId\") WHERE (\"byEmployeeId\" IS NOT NULL)",
			"CREATE INDEX discipline_event_employee_idx ON public.\"DisciplineEvent\" USING btree (\"employeeId\", \"createdAt\" DESC)",
			"CREATE UNIQUE INDEX discipline_event_unique_penalty ON public.\"DisciplineEvent\" USING btree (\"employeeId\", \"bookingId\", kind) WHERE (\"bookingId\" IS NOT NULL)",
		},
	})
}
