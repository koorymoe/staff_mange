// موديل جدول "BookingAssignment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingAssignment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "role", Type: "\"TechnicianRole\"", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "assignedById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "BookingAssignment_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingAssignment_bookingId_role_key", Def: "UNIQUE (\"bookingId\", role)"},
			{Name: "BookingAssignment_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "BookingAssignment_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"BookingAssignment_bookingId_idx\" ON public.\"BookingAssignment\" USING btree (\"bookingId\")",
			"CREATE INDEX \"BookingAssignment_employeeId_idx\" ON public.\"BookingAssignment\" USING btree (\"employeeId\")",
		},
	})
}
