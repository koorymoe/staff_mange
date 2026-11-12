// موديل جدول "SupplierSpecialty" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SupplierSpecialty",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "order", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "SupplierSpecialty_name_key", Def: "UNIQUE (name)"},
			{Name: "SupplierSpecialty_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
