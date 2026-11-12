// موديل جدول "ProcurementItem" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProcurementItem",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "requestId", Type: "text", NotNull: true},
			{Name: "productName", Type: "text", NotNull: true},
			{Name: "quantity", Type: "integer", NotNull: true},
			{Name: "unitPrice", Type: "double precision"},
			{Name: "totalPrice", Type: "double precision"},
			{Name: "fulfilled", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "model", Type: "text"},
			{Name: "spec", Type: "text"},
			{Name: "itemNotes", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "ProcurementItem_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProcurementItem_requestId_fkey", Def: "FOREIGN KEY (\"requestId\") REFERENCES \"ProcurementRequest\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"ProcurementItem_requestId_idx\" ON public.\"ProcurementItem\" USING btree (\"requestId\")",
		},
	})
}
