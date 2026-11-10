// موديل جدول "AttendanceClaim" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AttendanceClaim",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "attendanceId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "autoAt", Type: "timestamp with time zone", NotNull: true},
			{Name: "claimedUntil", Type: "timestamp with time zone"},
			{Name: "evidenceAt", Type: "timestamp with time zone"},
			{Name: "evidence", Type: "text"},
			{Name: "note", Type: "text"},
			{Name: "status", Type: "text", NotNull: true},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AttendanceClaim_attendanceId_fkey", Def: "FOREIGN KEY (\"attendanceId\") REFERENCES \"Attendance\"(id) ON DELETE CASCADE"},
			{Name: "AttendanceClaim_attendanceId_key", Def: "UNIQUE (\"attendanceId\")"},
			{Name: "AttendanceClaim_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id)"},
			{Name: "AttendanceClaim_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "AttendanceClaim_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AttendanceClaim_status_idx\" ON public.\"AttendanceClaim\" USING btree (status, \"createdAt\")",
		},
	})
}
