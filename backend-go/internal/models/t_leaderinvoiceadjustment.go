// موديل جدول "LeaderInvoiceAdjustment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "LeaderInvoiceAdjustment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "invoiceId", Type: "text", NotNull: true},
			{Name: "oldExecutionCost", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "newExecutionCost", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "oldMaterialsTotal", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "newMaterialsTotal", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "oldDiscountValue", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "newDiscountValue", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "oldNetTotal", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "newNetTotal", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "adjustedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "LeaderInvoiceAdjustment_adjustedById_fkey", Def: "FOREIGN KEY (\"adjustedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "LeaderInvoiceAdjustment_invoiceId_fkey", Def: "FOREIGN KEY (\"invoiceId\") REFERENCES \"LeaderInvoice\"(id) ON DELETE CASCADE"},
			{Name: "LeaderInvoiceAdjustment_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"LeaderInvoiceAdjustment_invoice_idx\" ON public.\"LeaderInvoiceAdjustment\" USING btree (\"invoiceId\", \"createdAt\" DESC)",
		},
	})
}
