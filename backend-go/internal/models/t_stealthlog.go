// موديل جدول "StealthLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "StealthLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "ownerId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "method", Type: "text", NotNull: true},
			{Name: "path", Type: "text", NotNull: true},
			{Name: "status", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "StealthLog_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"StealthLog_employee_idx\" ON public.\"StealthLog\" USING btree (\"employeeId\", \"createdAt\" DESC)",
		},
	})
}
