// موديل جدول "ServiceApp" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ServiceApp",
		Columns: []Column{
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "imageUrl", Type: "text"},
			{Name: "description", Type: "text"},
			{Name: "features", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "priceFrom", Type: "numeric"},
			{Name: "visible", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "ord", Type: "integer", NotNull: true, Default: "100"},
			{Name: "color", Type: "text"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ServiceApp_pkey", Def: "PRIMARY KEY (\"serviceId\")"},
			{Name: "ServiceApp_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
		},
	})
}
