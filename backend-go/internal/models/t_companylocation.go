// موديل جدول "CompanyLocation" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CompanyLocation",
		Columns: []Column{
			{Name: "id", Type: "integer", NotNull: true, Default: "1"},
			{Name: "lat", Type: "double precision", NotNull: true},
			{Name: "lng", Type: "double precision", NotNull: true},
			{Name: "radiusM", Type: "integer", NotNull: true, Default: "300"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "CompanyLocation_id_check", Def: "CHECK ((id = 1))"},
			{Name: "CompanyLocation_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
