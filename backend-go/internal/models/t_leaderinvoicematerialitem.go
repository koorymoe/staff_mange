// موديل جدول "LeaderInvoiceMaterialItem" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "LeaderInvoiceMaterialItem",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "leaderInvoiceId", Type: "text", NotNull: true},
			{Name: "materialId", Type: "text"},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "quantity", Type: "numeric", NotNull: true, Default: "1"},
			{Name: "unitPrice", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "profitPerUnit", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "lineTotal", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "LeaderInvoiceMaterialItem_leaderInvoiceId_fkey", Def: "FOREIGN KEY (\"leaderInvoiceId\") REFERENCES \"LeaderInvoice\"(id) ON DELETE CASCADE"},
			{Name: "LeaderInvoiceMaterialItem_materialId_fkey", Def: "FOREIGN KEY (\"materialId\") REFERENCES \"Material\"(id)"},
			{Name: "LeaderInvoiceMaterialItem_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"LeaderInvoiceMaterialItem_leaderInvoiceId_idx\" ON public.\"LeaderInvoiceMaterialItem\" USING btree (\"leaderInvoiceId\")",
		},
	})
}
