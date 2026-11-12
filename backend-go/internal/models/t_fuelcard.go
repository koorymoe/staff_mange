// موديل جدول "FuelCard" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "FuelCard",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "number", Type: "text", NotNull: true},
			{Name: "provider", Type: "text"},
			{Name: "vehicleId", Type: "text"},
			{Name: "holderEmployeeId", Type: "text"},
			{Name: "monthlyLimit", Type: "numeric"},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "notes", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "FuelCard_holderEmployeeId_fkey", Def: "FOREIGN KEY (\"holderEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "FuelCard_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "FuelCard_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE SET NULL"},
		},
	})
}
