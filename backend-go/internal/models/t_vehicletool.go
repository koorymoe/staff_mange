// موديل جدول "VehicleTool" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleTool",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "barcode", Type: "text"},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "status", Type: "\"ToolStatus\"", NotNull: true, Default: "'AVAILABLE'::\"ToolStatus\""},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "quantity", Type: "integer", NotNull: true, Default: "1"},
		},
		Constraints: []Constraint{
			{Name: "VehicleTool_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleTool_vehicleId_idx\" ON public.\"VehicleTool\" USING btree (\"vehicleId\")",
		},
	})
}
