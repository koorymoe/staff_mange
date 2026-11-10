// موديل جدول "QuotationItem" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "QuotationItem",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "quotationId", Type: "text", NotNull: true},
			{Name: "productName", Type: "text", NotNull: true},
			{Name: "unit", Type: "text"},
			{Name: "quantity", Type: "integer", NotNull: true},
			{Name: "unitPrice", Type: "double precision", NotNull: true},
			{Name: "totalPrice", Type: "double precision", NotNull: true},
			{Name: "imageBase64", Type: "text"},
			{Name: "sortIndex", Type: "integer", NotNull: true, Default: "0"},
			{Name: "modelName", Type: "text"},
			{Name: "specs", Type: "text"},
			{Name: "category", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "QuotationItem_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "QuotationItem_quotationId_fkey", Def: "FOREIGN KEY (\"quotationId\") REFERENCES \"Quotation\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"QuotationItem_order_idx\" ON public.\"QuotationItem\" USING btree (\"quotationId\", \"sortIndex\")",
			"CREATE INDEX \"QuotationItem_quotationId_idx\" ON public.\"QuotationItem\" USING btree (\"quotationId\")",
		},
	})
}
