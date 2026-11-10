// موديل جدول "BookingToolCheck" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingToolCheck",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "missingItems", Type: "text"},
			{Name: "checkedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "BookingToolCheck_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingToolCheck_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "BookingToolCheck_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"BookingToolCheck_bookingId_idx\" ON public.\"BookingToolCheck\" USING btree (\"bookingId\")",
			"CREATE INDEX \"BookingToolCheck_employeeId_idx\" ON public.\"BookingToolCheck\" USING btree (\"employeeId\")",
		},
	})
}
