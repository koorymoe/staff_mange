// موديل جدول "CoordinationAlert" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CoordinationAlert",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "coordinatorId", Type: "text"},
			{Name: "coordinatorName", Type: "text"},
			{Name: "reason", Type: "text"},
			{Name: "byEmployeeId", Type: "text"},
			{Name: "byName", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "resolvedAt", Type: "timestamp without time zone"},
			{Name: "resolvedById", Type: "text"},
			{Name: "resolvedByName", Type: "text"},
			{Name: "resolveNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "CoordinationAlert_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "CoordinationAlert_byEmployeeId_fkey", Def: "FOREIGN KEY (\"byEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "CoordinationAlert_coordinatorId_fkey", Def: "FOREIGN KEY (\"coordinatorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "CoordinationAlert_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "CoordinationAlert_resolvedById_fkey", Def: "FOREIGN KEY (\"resolvedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"CoordinationAlert_booking_time_idx\" ON public.\"CoordinationAlert\" USING btree (\"bookingId\", \"createdAt\" DESC)",
		},
	})
}
