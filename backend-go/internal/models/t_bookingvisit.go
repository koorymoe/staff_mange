// موديل جدول "BookingVisit" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingVisit",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "visitNumber", Type: "integer", NotNull: true},
			{Name: "outcome", Type: "text", NotNull: true},
			{Name: "percentDone", Type: "integer"},
			{Name: "progressReportId", Type: "text"},
			{Name: "scheduledAt", Type: "timestamp with time zone"},
			{Name: "occurredAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "BookingVisit_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingVisit_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "BookingVisit_progressReportId_fkey", Def: "FOREIGN KEY (\"progressReportId\") REFERENCES \"BookingProgressReport\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"BookingVisit_booking_number_key\" ON public.\"BookingVisit\" USING btree (\"bookingId\", \"visitNumber\")",
			"CREATE INDEX \"BookingVisit_occurred_idx\" ON public.\"BookingVisit\" USING btree (\"occurredAt\" DESC)",
		},
	})
}
