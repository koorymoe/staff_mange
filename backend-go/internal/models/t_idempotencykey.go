// موديل جدول "IdempotencyKey" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "IdempotencyKey",
		Columns: []Column{
			{Name: "key", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "method", Type: "text", NotNull: true},
			{Name: "path", Type: "text", NotNull: true},
			{Name: "status", Type: "integer"},
			{Name: "response", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "IdempotencyKey_pkey", Def: "PRIMARY KEY (key, \"employeeId\")"},
		},
		Indexes: []string{
			"CREATE INDEX \"IdempotencyKey_created_idx\" ON public.\"IdempotencyKey\" USING btree (\"createdAt\")",
		},
	})
}
