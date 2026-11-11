// موديل جدول "VehicleLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "type", Type: "text", NotNull: true},
			{Name: "performedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "nextDueAt", Type: "timestamp without time zone"},
			{Name: "odometer", Type: "integer"},
			{Name: "cost", Type: "double precision"},
			{Name: "notes", Type: "text"},
			{Name: "recordedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "nextDueOdometer", Type: "integer"},
			{Name: "liters", Type: "double precision"},
			{Name: "filledByEmployeeId", Type: "text"},
			{Name: "receiptNumber", Type: "text"},
			{Name: "stationName", Type: "text"},
			{Name: "receiptPhotoBase64", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "VehicleLog_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleLog_recordedById_fkey", Def: "FOREIGN KEY (\"recordedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "VehicleLog_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleLog_filledBy_performedAt_idx\" ON public.\"VehicleLog\" USING btree (\"filledByEmployeeId\", \"performedAt\")",
			"CREATE INDEX \"VehicleLog_vehicleId_idx\" ON public.\"VehicleLog\" USING btree (\"vehicleId\")",
		},
	})
}
