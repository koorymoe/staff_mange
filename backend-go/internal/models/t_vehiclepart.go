// موديل جدول "VehiclePart" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehiclePart",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "partType", Type: "text", NotNull: true},
			{Name: "installedAt", Type: "timestamp without time zone", NotNull: true},
			{Name: "installedOdometer", Type: "integer", NotNull: true},
			{Name: "expectedLifespanKm", Type: "integer"},
			{Name: "expectedLifespanMonths", Type: "integer"},
			{Name: "notes", Type: "text"},
			{Name: "replacedAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "cost", Type: "double precision"},
		},
		Constraints: []Constraint{
			{Name: "VehiclePart_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehiclePart_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehiclePart_vehicleId_idx\" ON public.\"VehiclePart\" USING btree (\"vehicleId\")",
		},
	})
}
