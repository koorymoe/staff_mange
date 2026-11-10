// موديل جدول "MaintenanceContract" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MaintenanceContract",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "intervalMonths", Type: "integer", NotNull: true},
			{Name: "nextVisitDate", Type: "date", NotNull: true},
			{Name: "address", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "lastBookingId", Type: "text"},
			{Name: "lastGeneratedAt", Type: "timestamp with time zone"},
			{Name: "lastError", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "MaintenanceContract_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "MaintenanceContract_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE CASCADE"},
			{Name: "MaintenanceContract_intervalMonths_check", Def: "CHECK (((\"intervalMonths\" >= 1) AND (\"intervalMonths\" <= 24)))"},
			{Name: "MaintenanceContract_lastBookingId_fkey", Def: "FOREIGN KEY (\"lastBookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "MaintenanceContract_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MaintenanceContract_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"MaintenanceContract_due_idx\" ON public.\"MaintenanceContract\" USING btree (\"nextVisitDate\") WHERE active",
		},
	})
}
