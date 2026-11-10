// موديل جدول "VehicleDriver" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleDriver",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "licenseNumber", Type: "text"},
			{Name: "licenseExpiry", Type: "date"},
			{Name: "authorized", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "authorizedVehicles", Type: "text"},
			{Name: "pledgeSignedAt", Type: "date"},
			{Name: "pledgeNote", Type: "text"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "VehicleDriver_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "VehicleDriver_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
			{Name: "VehicleDriver_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
