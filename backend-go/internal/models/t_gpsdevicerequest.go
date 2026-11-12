// موديل جدول "GpsDeviceRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "GpsDeviceRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "adminId", Type: "text"},
			{Name: "purchaseType", Type: "\"GpsPurchaseType\"", NotNull: true},
			{Name: "subscriptionType", Type: "\"GpsSubscriptionType\"", NotNull: true},
			{Name: "subscriptionStart", Type: "timestamp without time zone"},
			{Name: "subscriptionEnd", Type: "timestamp without time zone"},
			{Name: "subscriptionStatus", Type: "\"GpsSubscriptionStatus\"", NotNull: true, Default: "'ACTIVE'::\"GpsSubscriptionStatus\""},
			{Name: "status", Type: "\"GpsRequestStatus\"", NotNull: true, Default: "'PENDING'::\"GpsRequestStatus\""},
			{Name: "simCardId", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "isChecked", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "isActivated", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "isDelivered", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "invoicePhotoUrl", Type: "text"},
			{Name: "gpsNumber", Type: "text"},
			{Name: "residenceCardNumber", Type: "text"},
			{Name: "activationDate", Type: "timestamp without time zone"},
			{Name: "deliveredAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "scheduledAt", Type: "timestamp without time zone"},
			{Name: "assignedTechnicianId", Type: "text"},
			{Name: "credentialsMessage", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "GpsDeviceRequest_adminId_fkey", Def: "FOREIGN KEY (\"adminId\") REFERENCES \"Employee\"(id)"},
			{Name: "GpsDeviceRequest_assignedTechnicianId_fkey", Def: "FOREIGN KEY (\"assignedTechnicianId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "GpsDeviceRequest_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"GpsCustomer\"(id)"},
			{Name: "GpsDeviceRequest_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "GpsDeviceRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "GpsDeviceRequest_simCardId_fkey", Def: "FOREIGN KEY (\"simCardId\") REFERENCES \"SimCard\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"GpsDeviceRequest_assignedTechnicianId_idx\" ON public.\"GpsDeviceRequest\" USING btree (\"assignedTechnicianId\")",
			"CREATE INDEX \"GpsDeviceRequest_customerId_idx\" ON public.\"GpsDeviceRequest\" USING btree (\"customerId\")",
			"CREATE INDEX \"GpsDeviceRequest_employeeId_idx\" ON public.\"GpsDeviceRequest\" USING btree (\"employeeId\")",
			"CREATE INDEX \"GpsDeviceRequest_simCardId_idx\" ON public.\"GpsDeviceRequest\" USING btree (\"simCardId\")",
			"CREATE INDEX \"GpsDeviceRequest_status_idx\" ON public.\"GpsDeviceRequest\" USING btree (status)",
		},
	})
}
