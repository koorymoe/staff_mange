// موديل جدول "VehicleMission" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleMission",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "driverId", Type: "text", NotNull: true},
			{Name: "purpose", Type: "text", NotNull: true},
			{Name: "destination", Type: "text", NotNull: true},
			{Name: "startedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "endedAt", Type: "timestamp without time zone"},
			{Name: "startOdometer", Type: "integer", NotNull: true},
			{Name: "endOdometer", Type: "integer"},
			{Name: "distanceKm", Type: "integer"},
			{Name: "notes", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'IN_PROGRESS'::text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehicleMission_driverId_fkey", Def: "FOREIGN KEY (\"driverId\") REFERENCES \"Employee\"(id)"},
			{Name: "VehicleMission_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleMission_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleMission_driverId_idx\" ON public.\"VehicleMission\" USING btree (\"driverId\")",
			"CREATE INDEX \"VehicleMission_status_idx\" ON public.\"VehicleMission\" USING btree (status)",
			"CREATE INDEX \"VehicleMission_vehicleId_idx\" ON public.\"VehicleMission\" USING btree (\"vehicleId\")",
		},
	})
}
