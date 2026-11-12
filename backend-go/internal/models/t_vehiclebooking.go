// موديل جدول "VehicleBooking" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleBooking",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "requestedById", Type: "text", NotNull: true},
			{Name: "purpose", Type: "text", NotNull: true},
			{Name: "startAt", Type: "timestamp without time zone", NotNull: true},
			{Name: "endAt", Type: "timestamp without time zone", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "approvedById", Type: "text"},
			{Name: "rejectionReason", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "decidedAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "VehicleBooking_approvedById_fkey", Def: "FOREIGN KEY (\"approvedById\") REFERENCES \"Employee\"(id)"},
			{Name: "VehicleBooking_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleBooking_requestedById_fkey", Def: "FOREIGN KEY (\"requestedById\") REFERENCES \"Employee\"(id)"},
			{Name: "VehicleBooking_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleBooking_requestedById_idx\" ON public.\"VehicleBooking\" USING btree (\"requestedById\")",
			"CREATE INDEX \"VehicleBooking_status_idx\" ON public.\"VehicleBooking\" USING btree (status)",
			"CREATE INDEX \"VehicleBooking_vehicleId_idx\" ON public.\"VehicleBooking\" USING btree (\"vehicleId\")",
		},
	})
}
