// موديل جدول "ProcurementPriceAlert" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProcurementPriceAlert",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "procurementId", Type: "text"},
			{Name: "itemId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "productName", Type: "text", NotNull: true},
			{Name: "supplierId", Type: "text"},
			{Name: "unitPrice", Type: "double precision", NotNull: true},
			{Name: "quantity", Type: "integer", NotNull: true},
			{Name: "refSupplierId", Type: "text"},
			{Name: "refPrice", Type: "double precision", NotNull: true},
			{Name: "lossIqd", Type: "double precision", NotNull: true},
			{Name: "buyerId", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "reviewedAt", Type: "timestamp without time zone"},
			{Name: "reviewNote", Type: "text"},
			{Name: "expenseId", Type: "text"},
			{Name: "techSupplierId", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "ProcurementPriceAlert_expenseId_fkey", Def: "FOREIGN KEY (\"expenseId\") REFERENCES \"Expense\"(id) ON DELETE CASCADE"},
			{Name: "ProcurementPriceAlert_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProcurementPriceAlert_procurementId_fkey", Def: "FOREIGN KEY (\"procurementId\") REFERENCES \"ProcurementRequest\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX price_alert_once ON public.\"ProcurementPriceAlert\" USING btree (\"itemId\", kind)",
		},
	})
}
