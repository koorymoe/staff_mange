// موديل جدول "AgentGap" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AgentGap",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "agent", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "missing", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AgentGap_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AgentGap_kind_idx\" ON public.\"AgentGap\" USING btree (agent, kind, \"createdAt\" DESC)",
		},
	})
}
