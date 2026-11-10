// موديل جدول "VehicleIncident" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleIncident",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "type", Type: "text", NotNull: true},
			{Name: "description", Type: "text", NotNull: true},
			{Name: "responsibleEmployeeId", Type: "text"},
			{Name: "cost", Type: "double precision"},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "reportedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "resolvedAt", Type: "timestamp without time zone"},
			{Name: "location", Type: "text"},
			{Name: "driverId", Type: "text"},
			{Name: "peoplePresent", Type: "text"},
			{Name: "policeReportNumber", Type: "text"},
			{Name: "repairCost", Type: "double precision"},
		},
		Constraints: []Constraint{
			{Name: "VehicleIncident_driverId_fkey", Def: "FOREIGN KEY (\"driverId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "VehicleIncident_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleIncident_reportedById_fkey", Def: "FOREIGN KEY (\"reportedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "VehicleIncident_responsibleEmployeeId_fkey", Def: "FOREIGN KEY (\"responsibleEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "VehicleIncident_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleIncident_vehicleId_idx\" ON public.\"VehicleIncident\" USING btree (\"vehicleId\")",
		},
	})
}
