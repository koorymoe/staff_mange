// موديل جدول "GpsMaintenanceRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "GpsMaintenanceRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "adminId", Type: "text"},
			{Name: "problemDescription", Type: "text", NotNull: true},
			{Name: "status", Type: "\"GpsMaintenanceStatus\"", NotNull: true, Default: "'PENDING'::\"GpsMaintenanceStatus\""},
			{Name: "adminNotes", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "resolvedAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "GpsMaintenanceRequest_adminId_fkey", Def: "FOREIGN KEY (\"adminId\") REFERENCES \"Employee\"(id)"},
			{Name: "GpsMaintenanceRequest_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"GpsCustomer\"(id)"},
			{Name: "GpsMaintenanceRequest_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "GpsMaintenanceRequest_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"GpsMaintenanceRequest_customerId_idx\" ON public.\"GpsMaintenanceRequest\" USING btree (\"customerId\")",
			"CREATE INDEX \"GpsMaintenanceRequest_employeeId_idx\" ON public.\"GpsMaintenanceRequest\" USING btree (\"employeeId\")",
		},
	})
}
