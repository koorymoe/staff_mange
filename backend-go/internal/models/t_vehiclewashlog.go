// موديل جدول "VehicleWashLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleWashLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "washDate", Type: "date", NotNull: true},
			{Name: "washedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehicleWashLog_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "VehicleWashLog_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleWashLog_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"VehicleWashLog_vehicle_day\" ON public.\"VehicleWashLog\" USING btree (\"vehicleId\", \"washDate\")",
		},
	})
}
