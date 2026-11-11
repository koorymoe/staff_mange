// موديل جدول "StockIntake" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "StockIntake",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "toolId", Type: "text", NotNull: true},
			{Name: "quantity", Type: "integer", NotNull: true},
			{Name: "unitPrice", Type: "numeric(14,2)"},
			{Name: "supplier", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "StockIntake_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "StockIntake_toolId_fkey", Def: "FOREIGN KEY (\"toolId\") REFERENCES \"OnDemandTool\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"StockIntake_tool_idx\" ON public.\"StockIntake\" USING btree (\"toolId\", \"createdAt\" DESC)",
		},
	})
}
