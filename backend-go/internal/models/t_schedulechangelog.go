// موديل جدول "ScheduleChangeLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ScheduleChangeLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "changedById", Type: "text", NotNull: true},
			{Name: "oldTime", Type: "timestamp without time zone"},
			{Name: "newTime", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "reason", Type: "text"},
			{Name: "kind", Type: "text", NotNull: true, Default: "'SCHEDULE'::text"},
		},
		Constraints: []Constraint{
			{Name: "ScheduleChangeLog_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "ScheduleChangeLog_changedById_fkey", Def: "FOREIGN KEY (\"changedById\") REFERENCES \"Employee\"(id)"},
			{Name: "ScheduleChangeLog_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ScheduleChangeLog_bookingId_idx\" ON public.\"ScheduleChangeLog\" USING btree (\"bookingId\")",
			"CREATE INDEX \"ScheduleChangeLog_changedById_idx\" ON public.\"ScheduleChangeLog\" USING btree (\"changedById\")",
		},
	})
}
