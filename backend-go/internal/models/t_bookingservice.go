// موديل جدول "BookingService" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingService",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "BookingService_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingService_bookingId_serviceId_key", Def: "UNIQUE (\"bookingId\", \"serviceId\")"},
			{Name: "BookingService_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "BookingService_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"BookingService_booking_idx\" ON public.\"BookingService\" USING btree (\"bookingId\")",
		},
	})
}
