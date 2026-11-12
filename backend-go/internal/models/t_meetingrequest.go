// موديل جدول "MeetingRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MeetingRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "guestName", Type: "text"},
			{Name: "preferredAt", Type: "timestamp with time zone"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "scheduledAt", Type: "timestamp with time zone"},
			{Name: "note", Type: "text"},
			{Name: "decidedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "MeetingRequest_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "MeetingRequest_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "MeetingRequest_kind_check", Def: "CHECK ((kind = ANY (ARRAY['MEETING'::text, 'GUEST'::text])))"},
			{Name: "MeetingRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MeetingRequest_status_check", Def: "CHECK ((status = ANY (ARRAY['PENDING'::text, 'SCHEDULED'::text, 'DECLINED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"MeetingRequest_emp_idx\" ON public.\"MeetingRequest\" USING btree (\"employeeId\", \"createdAt\" DESC)",
			"CREATE INDEX \"MeetingRequest_pending_idx\" ON public.\"MeetingRequest\" USING btree (\"createdAt\") WHERE (status = 'PENDING'::text)",
		},
	})
}
