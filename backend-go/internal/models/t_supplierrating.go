// موديل جدول "SupplierRating" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SupplierRating",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "supplierId", Type: "text", NotNull: true},
			{Name: "value", Type: "integer", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "ratedById", Type: "text", NotNull: true},
			{Name: "ratedByName", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "SupplierRating_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SupplierRating_ratedById_fkey", Def: "FOREIGN KEY (\"ratedById\") REFERENCES \"Employee\"(id)"},
			{Name: "SupplierRating_supplierId_fkey", Def: "FOREIGN KEY (\"supplierId\") REFERENCES \"Supplier\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"SupplierRating_supplierId_idx\" ON public.\"SupplierRating\" USING btree (\"supplierId\")",
		},
	})
}
