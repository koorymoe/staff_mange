// موديل جدول "TrainingMaterial" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingMaterial",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "url", Type: "text", NotNull: true},
			{Name: "type", Type: "text", NotNull: true, Default: "'VIDEO'::text"},
			{Name: "order", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "TrainingMaterial_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "TrainingMaterial_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
		},
	})
}
