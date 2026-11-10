// موديل جدول "ProcurementQuote" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProcurementQuote",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "procurementId", Type: "text", NotNull: true},
			{Name: "supplierId", Type: "text", NotNull: true},
			{Name: "totalPrice", Type: "double precision", NotNull: true},
			{Name: "chosen", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "note", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ProcurementQuote_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProcurementQuote_procurementId_fkey", Def: "FOREIGN KEY (\"procurementId\") REFERENCES \"ProcurementRequest\"(id) ON DELETE CASCADE"},
			{Name: "ProcurementQuote_supplierId_fkey", Def: "FOREIGN KEY (\"supplierId\") REFERENCES \"Supplier\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX procurement_quote_req ON public.\"ProcurementQuote\" USING btree (\"procurementId\")",
		},
	})
}
