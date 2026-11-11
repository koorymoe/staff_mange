// موديل جدول "BookingRouteStop" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingRouteStop",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "lat", Type: "double precision", NotNull: true},
			{Name: "lng", Type: "double precision", NotNull: true},
			{Name: "startAt", Type: "timestamp without time zone", NotNull: true},
			{Name: "endAt", Type: "timestamp without time zone", NotNull: true},
			{Name: "minutes", Type: "integer", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "BookingRouteStop_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingRouteStop_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX route_stop_booking ON public.\"BookingRouteStop\" USING btree (\"bookingId\")",
		},
	})
}
