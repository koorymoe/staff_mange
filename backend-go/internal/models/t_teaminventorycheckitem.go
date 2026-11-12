// موديل جدول "TeamInventoryCheckItem" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TeamInventoryCheckItem",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "checkId", Type: "text", NotNull: true},
			{Name: "toolName", Type: "text", NotNull: true},
			{Name: "personRole", Type: "text", NotNull: true},
			{Name: "present", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "reason", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "TeamInventoryCheckItem_checkId_fkey", Def: "FOREIGN KEY (\"checkId\") REFERENCES \"TeamInventoryCheck\"(id) ON DELETE CASCADE"},
			{Name: "TeamInventoryCheckItem_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"TeamInventoryCheckItem_checkId_idx\" ON public.\"TeamInventoryCheckItem\" USING btree (\"checkId\")",
		},
	})
}
