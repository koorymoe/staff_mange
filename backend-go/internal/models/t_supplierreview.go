// موديل جدول "SupplierReview" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SupplierReview",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "supplierId", Type: "text", NotNull: true},
			{Name: "procurementId", Type: "text"},
			{Name: "price", Type: "integer", NotNull: true},
			{Name: "quality", Type: "integer", NotNull: true},
			{Name: "speed", Type: "integer", NotNull: true},
			{Name: "punctuality", Type: "integer", NotNull: true},
			{Name: "dealing", Type: "integer", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "ratedById", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "SupplierReview_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SupplierReview_procurementId_fkey", Def: "FOREIGN KEY (\"procurementId\") REFERENCES \"ProcurementRequest\"(id) ON DELETE SET NULL"},
			{Name: "SupplierReview_supplierId_fkey", Def: "FOREIGN KEY (\"supplierId\") REFERENCES \"Supplier\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX supplier_review_once ON public.\"SupplierReview\" USING btree (\"procurementId\", \"ratedById\") WHERE (\"procurementId\" IS NOT NULL)",
			"CREATE INDEX supplier_review_sup ON public.\"SupplierReview\" USING btree (\"supplierId\")",
		},
	})
}
