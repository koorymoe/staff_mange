// موديل جدول "InventoryCheck" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "InventoryCheck",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "complete", Type: "boolean", NotNull: true},
			{Name: "missingItems", Type: "text"},
			{Name: "checkedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "resolved", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "resolvedById", Type: "text"},
			{Name: "resolvedAt", Type: "timestamp without time zone"},
			{Name: "bookingId", Type: "text"},
			{Name: "priority", Type: "text"},
			{Name: "priorityById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "InventoryCheck_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "InventoryCheck_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "InventoryCheck_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "InventoryCheck_priorityById_fkey", Def: "FOREIGN KEY (\"priorityById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "InventoryCheck_priority_check", Def: "CHECK ((priority = ANY (ARRAY['URGENT'::text, 'HIGH'::text, 'NORMAL'::text])))"},
			{Name: "InventoryCheck_resolvedById_fkey", Def: "FOREIGN KEY (\"resolvedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"InventoryCheck_checkedAt_idx\" ON public.\"InventoryCheck\" USING btree (\"checkedAt\")",
			"CREATE INDEX \"InventoryCheck_employeeId_idx\" ON public.\"InventoryCheck\" USING btree (\"employeeId\")",
			"CREATE UNIQUE INDEX inventory_check_booking_employee_key ON public.\"InventoryCheck\" USING btree (\"bookingId\", \"employeeId\") WHERE (\"bookingId\" IS NOT NULL)",
			"CREATE INDEX inventory_check_booking_idx ON public.\"InventoryCheck\" USING btree (\"bookingId\")",
		},
	})
}
