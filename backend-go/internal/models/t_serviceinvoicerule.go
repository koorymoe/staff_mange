// موديل جدول "ServiceInvoiceRule" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-24 دازاي: «آخر خيار: شنو أقل سعر تريده يظهر بالفاتورة» — على الشغل بس (المواد فوگه).
func init() {
	register(Table{
		Name: "ServiceInvoiceRule",
		Columns: []Column{
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "minWork", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "minTiers", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"}, // (دازاي 11-28) درجات: أقل من 14 ← 14، أكثر ← 24…
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ServiceInvoiceRule_pkey", Def: "PRIMARY KEY (\"serviceId\")"},
			{Name: "ServiceInvoiceRule_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
			{Name: "ServiceInvoiceRule_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "service_invoice_rule_min_ok", Def: "CHECK ((\"minWork\" >= (0)::numeric))"},
		},
	})
}
