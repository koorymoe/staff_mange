// موديل جدول "Vehicle" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Vehicle",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "plateNumber", Type: "text", NotNull: true},
			{Name: "color", Type: "text"},
			{Name: "type", Type: "text"},
			{Name: "isActive", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "model", Type: "text"},
			{Name: "year", Type: "integer"},
			{Name: "chassisNumber", Type: "text"},
			{Name: "engineNumber", Type: "text"},
			{Name: "fuelType", Type: "text"},
			{Name: "currentOdometer", Type: "integer", NotNull: true, Default: "0"},
			{Name: "condition", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "Vehicle_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "Vehicle_plateNumber_key", Def: "UNIQUE (\"plateNumber\")"},
		},
	})
}
