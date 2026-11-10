// موديل جدول "LeaveRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "LeaveRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "startDate", Type: "date", NotNull: true},
			{Name: "endDate", Type: "date", NotNull: true},
			{Name: "reason", Type: "text"},
			{Name: "route", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "decisionNote", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "preliminaryById", Type: "text"},
			{Name: "preliminaryAt", Type: "timestamp with time zone"},
			{Name: "preliminaryNote", Type: "text"},
			{Name: "kind", Type: "text", NotNull: true, Default: "'REGULAR'::text"},
		},
		Constraints: []Constraint{
			{Name: "LeaveRequest_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id)"},
			{Name: "LeaveRequest_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "LeaveRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "LeaveRequest_preliminaryById_fkey", Def: "FOREIGN KEY (\"preliminaryById\") REFERENCES \"Employee\"(id)"},
			{Name: "leave_request_kind_valid", Def: "CHECK ((kind = ANY (ARRAY['REGULAR'::text, 'URGENT'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"LeaveRequest_employee_idx\" ON public.\"LeaveRequest\" USING btree (\"employeeId\", \"createdAt\" DESC)",
			"CREATE INDEX \"LeaveRequest_pending_idx\" ON public.\"LeaveRequest\" USING btree (status, route, \"startDate\")",
		},
	})
}
