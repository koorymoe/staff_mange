// موديل جدول "ServiceInvoiceOption" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-24 دازاي: «نضيف بالإعدادات فاتورة لكل خدمة — خانات (تسليك كيبل بسعر المتر، غسل لوح بسعر اللوح…)».
// تجريبي: ما يأثر على الفواتير الحقيقية لحد ما دازاي يطلقه.
func init() {
	register(Table{
		Name: "ServiceInvoiceOption",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "label", Type: "text", NotNull: true},
			{Name: "unit", Type: "text", NotNull: true, Default: "'قطعة'::text"},
			{Name: "unitPrice", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "sortOrder", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ServiceInvoiceOption_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ServiceInvoiceOption_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
			{Name: "service_invoice_option_price_ok", Def: "CHECK ((\"unitPrice\" >= (0)::numeric))"},
		},
		Indexes: []string{
			"CREATE INDEX \"ServiceInvoiceOption_service_idx\" ON public.\"ServiceInvoiceOption\" USING btree (\"serviceId\", \"sortOrder\")",
		},
	})
}
