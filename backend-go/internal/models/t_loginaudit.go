// موديل جدول "LoginAudit" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "LoginAudit",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "username", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "success", Type: "boolean", NotNull: true},
			{Name: "ipAddress", Type: "text"},
			{Name: "userAgent", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "LoginAudit_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "LoginAudit_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"LoginAudit_createdAt_idx\" ON public.\"LoginAudit\" USING btree (\"createdAt\")",
			"CREATE INDEX \"LoginAudit_employeeId_idx\" ON public.\"LoginAudit\" USING btree (\"employeeId\", \"createdAt\")",
		},
	})
}
