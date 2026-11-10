// موديل جدول "OnDemandTool" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OnDemandTool",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "barcode", Type: "text", NotNull: true},
			{Name: "totalQuantity", Type: "integer", NotNull: true},
			{Name: "availableQuantity", Type: "integer", NotNull: true},
			{Name: "status", Type: "\"ToolStatus\"", NotNull: true, Default: "'AVAILABLE'::\"ToolStatus\""},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "OnDemandTool_barcode_key", Def: "UNIQUE (barcode)"},
			{Name: "OnDemandTool_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
