// موديل جدول "BookingProgressReport" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingProgressReport",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "dayNumber", Type: "integer", NotNull: true, Default: "1"},
			{Name: "reportedById", Type: "text"},
			{Name: "workDone", Type: "text", NotNull: true},
			{Name: "remainingWork", Type: "text", NotNull: true},
			{Name: "percentDone", Type: "integer", NotNull: true, Default: "0"},
			{Name: "blockers", Type: "text"},
			{Name: "materialsUsed", Type: "text"},
			{Name: "crewSnapshot", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "sheet", Type: "jsonb"},
			{Name: "reviewedAt", Type: "timestamp without time zone"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "BookingProgressReport_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingProgressReport_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "BookingProgressReport_reportedById_fkey", Def: "FOREIGN KEY (\"reportedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "BookingProgressReport_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"BookingProgressReport_booking_idx\" ON public.\"BookingProgressReport\" USING btree (\"bookingId\", \"dayNumber\")",
		},
	})
}
