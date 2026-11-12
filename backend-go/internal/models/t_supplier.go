// موديل جدول "Supplier" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Supplier",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "companyName", Type: "text", NotNull: true},
			{Name: "ownerName", Type: "text", NotNull: true},
			{Name: "phone", Type: "text", NotNull: true},
			{Name: "lat", Type: "double precision"},
			{Name: "lng", Type: "double precision"},
			{Name: "isMaterialSupplier", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "isContractor", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "traderTypes", Type: "text[]", Default: "'{}'::text[]"},
			{Name: "notes", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "address", Type: "text"},
			{Name: "locationUrl", Type: "text"},
			{Name: "isCompetitor", Type: "boolean", NotNull: true, Default: "false"},
		},
		Constraints: []Constraint{
			{Name: "Supplier_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
