// موديل جدول "BookingAfterInventory" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingAfterInventory",
		Columns: []Column{
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "complete", Type: "boolean", NotNull: true},
			{Name: "missingItems", Type: "text"},
			{Name: "checkedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "BookingAfterInventory_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingAfterInventory_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "BookingAfterInventory_pkey", Def: "PRIMARY KEY (\"bookingId\", \"employeeId\")"},
		},
		Indexes: []string{
			"CREATE INDEX \"BookingAfterInventory_emp_idx\" ON public.\"BookingAfterInventory\" USING btree (\"employeeId\", \"checkedAt\" DESC)",
		},
	})
}
