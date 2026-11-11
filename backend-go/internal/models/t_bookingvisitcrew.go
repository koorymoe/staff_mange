// موديل جدول "BookingVisitCrew" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingVisitCrew",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "visitId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "isLeader", Type: "boolean", NotNull: true, Default: "false"},
		},
		Constraints: []Constraint{
			{Name: "BookingVisitCrew_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "BookingVisitCrew_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "BookingVisitCrew_visitId_fkey", Def: "FOREIGN KEY (\"visitId\") REFERENCES \"BookingVisit\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"BookingVisitCrew_employee_idx\" ON public.\"BookingVisitCrew\" USING btree (\"employeeId\")",
			"CREATE UNIQUE INDEX \"BookingVisitCrew_visit_employee_key\" ON public.\"BookingVisitCrew\" USING btree (\"visitId\", \"employeeId\")",
		},
	})
}
