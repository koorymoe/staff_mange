// موديل جدول "AgentEvalRun" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AgentEvalRun",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "agent", Type: "text", NotNull: true},
			{Name: "build", Type: "text", NotNull: true},
			{Name: "total", Type: "integer", NotNull: true},
			{Name: "passed", Type: "integer", NotNull: true},
			{Name: "details", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AgentEvalRun_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AgentEvalRun_agent_idx\" ON public.\"AgentEvalRun\" USING btree (agent, \"createdAt\" DESC)",
		},
	})
}
