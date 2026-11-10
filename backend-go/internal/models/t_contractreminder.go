// موديل جدول "ContractReminder" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ContractReminder",
		Columns: []Column{
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "forDate", Type: "date", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ContractReminder_pkey", Def: "PRIMARY KEY (kind, \"employeeId\", \"forDate\")"},
		},
	})
}
