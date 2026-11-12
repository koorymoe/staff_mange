// موديل جدول "VehiclePhoto" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehiclePhoto",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "url", Type: "text", NotNull: true},
			{Name: "caption", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehiclePhoto_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehiclePhoto_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehiclePhoto_vehicleId_idx\" ON public.\"VehiclePhoto\" USING btree (\"vehicleId\")",
		},
	})
}
