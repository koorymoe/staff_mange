// موديل جدول "LegacyPendingReset" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "LegacyPendingReset",
		Columns: []Column{
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "oldConfirmedAt", Type: "timestamp with time zone"},
			{Name: "oldContactedAt", Type: "timestamp with time zone"},
			{Name: "oldStatus", Type: "text", NotNull: true},
			{Name: "resetAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "LegacyPendingReset_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "LegacyPendingReset_pkey", Def: "PRIMARY KEY (\"bookingId\")"},
		},
	})
}
