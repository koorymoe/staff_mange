// موديل جدول "VehicleMissionPassenger" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleMissionPassenger",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "missionId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "VehicleMissionPassenger_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "VehicleMissionPassenger_missionId_employeeId_key", Def: "UNIQUE (\"missionId\", \"employeeId\")"},
			{Name: "VehicleMissionPassenger_missionId_fkey", Def: "FOREIGN KEY (\"missionId\") REFERENCES \"VehicleMission\"(id) ON DELETE CASCADE"},
			{Name: "VehicleMissionPassenger_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
