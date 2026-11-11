// موديل جدول "SecretaryAction" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SecretaryAction",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "request", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "summary", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "fields", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PROPOSED'::text"},
			{Name: "result", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "SecretaryAction_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "SecretaryAction_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SecretaryAction_status_check", Def: "CHECK ((status = ANY (ARRAY['PROPOSED'::text, 'SENT'::text, 'CANCELLED'::text, 'FAILED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"SecretaryAction_emp_idx\" ON public.\"SecretaryAction\" USING btree (\"employeeId\", \"createdAt\" DESC)",
		},
	})
}
