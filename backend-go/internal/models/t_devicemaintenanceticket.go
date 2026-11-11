// موديل جدول "DeviceMaintenanceTicket" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DeviceMaintenanceTicket",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "appointmentDate", Type: "timestamp without time zone"},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "deviceTypeName", Type: "text", NotNull: true},
			{Name: "problem", Type: "text", NotNull: true},
			{Name: "deviceSerial", Type: "text"},
			{Name: "receivedAt", Type: "timestamp without time zone"},
			{Name: "deliveredAt", Type: "timestamp without time zone"},
			{Name: "invoiceNumber", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "DeviceMaintenanceTicket_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id)"},
			{Name: "DeviceMaintenanceTicket_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "DeviceMaintenanceTicket_invoiceNumber_key", Def: "UNIQUE (\"invoiceNumber\")"},
			{Name: "DeviceMaintenanceTicket_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"DeviceMaintenanceTicket_customerId_idx\" ON public.\"DeviceMaintenanceTicket\" USING btree (\"customerId\")",
		},
	})
}
