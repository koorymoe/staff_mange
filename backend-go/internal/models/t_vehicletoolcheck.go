// موديل جدول "VehicleToolCheck" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleToolCheck",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "missionId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "missingToolNames", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehicleToolCheck_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "VehicleToolCheck_missionId_fkey", Def: "FOREIGN KEY (\"missionId\") REFERENCES \"VehicleMission\"(id) ON DELETE CASCADE"},
			{Name: "VehicleToolCheck_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleToolCheck_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleToolCheck_missionId_idx\" ON public.\"VehicleToolCheck\" USING btree (\"missionId\")",
			"CREATE INDEX \"VehicleToolCheck_vehicleId_idx\" ON public.\"VehicleToolCheck\" USING btree (\"vehicleId\")",
		},
	})
}
