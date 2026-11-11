// موديل جدول "PermissionAuditLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PermissionAuditLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "actorId", Type: "text"},
			{Name: "action", Type: "text", NotNull: true},
			{Name: "permission", Type: "text", NotNull: true},
			{Name: "reason", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "PermissionAuditLog_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"PermissionAuditLog_emp\" ON public.\"PermissionAuditLog\" USING btree (\"employeeId\", \"createdAt\" DESC)",
		},
	})
}
