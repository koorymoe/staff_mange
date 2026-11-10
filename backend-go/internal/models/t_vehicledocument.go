// موديل جدول "VehicleDocument" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleDocument",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "documentType", Type: "text", NotNull: true},
			{Name: "documentNumber", Type: "text"},
			{Name: "issueDate", Type: "timestamp without time zone"},
			{Name: "expiryDate", Type: "timestamp without time zone"},
			{Name: "fileUrl", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehicleDocument_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleDocument_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleDocument_vehicleId_idx\" ON public.\"VehicleDocument\" USING btree (\"vehicleId\")",
		},
	})
}
