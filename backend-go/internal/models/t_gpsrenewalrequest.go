// موديل جدول "GpsRenewalRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "GpsRenewalRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "deviceRequestId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "adminId", Type: "text"},
			{Name: "subscriptionType", Type: "\"GpsSubscriptionType\"", NotNull: true},
			{Name: "newEndDate", Type: "timestamp without time zone"},
			{Name: "status", Type: "\"GpsRequestStatus\"", NotNull: true, Default: "'PENDING'::\"GpsRequestStatus\""},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "GpsRenewalRequest_adminId_fkey", Def: "FOREIGN KEY (\"adminId\") REFERENCES \"Employee\"(id)"},
			{Name: "GpsRenewalRequest_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"GpsCustomer\"(id)"},
			{Name: "GpsRenewalRequest_deviceRequestId_fkey", Def: "FOREIGN KEY (\"deviceRequestId\") REFERENCES \"GpsDeviceRequest\"(id)"},
			{Name: "GpsRenewalRequest_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "GpsRenewalRequest_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"GpsRenewalRequest_customerId_idx\" ON public.\"GpsRenewalRequest\" USING btree (\"customerId\")",
			"CREATE INDEX \"GpsRenewalRequest_deviceRequestId_idx\" ON public.\"GpsRenewalRequest\" USING btree (\"deviceRequestId\")",
		},
	})
}
