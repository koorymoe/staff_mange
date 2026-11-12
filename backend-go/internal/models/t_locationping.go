// موديل جدول "LocationPing" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "LocationPing",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "latitude", Type: "double precision", NotNull: true},
			{Name: "longitude", Type: "double precision", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "accuracy", Type: "double precision"},
		},
		Constraints: []Constraint{
			{Name: "LocationPing_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "LocationPing_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "LocationPing_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"LocationPing_bookingId_idx\" ON public.\"LocationPing\" USING btree (\"bookingId\") WHERE (\"bookingId\" IS NOT NULL)",
			"CREATE INDEX \"LocationPing_createdAt_idx\" ON public.\"LocationPing\" USING btree (\"createdAt\")",
			"CREATE INDEX \"LocationPing_employeeId_idx\" ON public.\"LocationPing\" USING btree (\"employeeId\", \"createdAt\")",
		},
	})
}
