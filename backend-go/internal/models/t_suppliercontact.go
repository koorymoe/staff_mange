// موديل جدول "SupplierContact" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SupplierContact",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "supplierKind", Type: "text", NotNull: true},
			{Name: "supplierId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "byId", Type: "text"},
			{Name: "at", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "SupplierContact_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX supplier_contact_sup ON public.\"SupplierContact\" USING btree (\"supplierKind\", \"supplierId\", at DESC)",
		},
	})
}
