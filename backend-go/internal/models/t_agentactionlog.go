// موديل جدول "AgentActionLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AgentActionLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "agent", Type: "text", NotNull: true},
			{Name: "tool", Type: "text", NotNull: true},
			{Name: "actorId", Type: "text"},
			{Name: "input", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "result", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "failed", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AgentActionLog_actorId_fkey", Def: "FOREIGN KEY (\"actorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "AgentActionLog_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AgentActionLog_at\" ON public.\"AgentActionLog\" USING btree (\"createdAt\" DESC)",
		},
	})
}
