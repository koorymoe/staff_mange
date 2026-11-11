// موديل جدول "AiWatch" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiWatch",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "ownerId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "noteId", Type: "text"},
			{Name: "concern", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "baseline", Type: "jsonb", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "checkedAt", Type: "timestamp with time zone"},
			{Name: "closedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "AiWatch_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "AiWatch_ownerId_fkey", Def: "FOREIGN KEY (\"ownerId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "AiWatch_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "AiWatch_status_check", Def: "CHECK ((status = ANY (ARRAY['OPEN'::text, 'IMPROVED'::text, 'WORSE'::text, 'CLOSED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"AiWatch_open_idx\" ON public.\"AiWatch\" USING btree (status, \"createdAt\")",
		},
	})
}
