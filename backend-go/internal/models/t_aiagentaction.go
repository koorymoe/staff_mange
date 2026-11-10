// موديل جدول "AiAgentAction" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiAgentAction",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "agent", Type: "text", NotNull: true},
			{Name: "conversationId", Type: "text"},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "summary", Type: "text", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "targetType", Type: "text"},
			{Name: "targetId", Type: "text"},
			{Name: "payload", Type: "jsonb"},
			{Name: "sandbox", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "undoneAt", Type: "timestamp with time zone"},
			{Name: "undoneById", Type: "text"},
			{Name: "undoReason", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiAgentAction_conversationId_fkey", Def: "FOREIGN KEY (\"conversationId\") REFERENCES \"WaConversation\"(id) ON DELETE SET NULL"},
			{Name: "AiAgentAction_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "AiAgentAction_undoneById_fkey", Def: "FOREIGN KEY (\"undoneById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"AiAgentAction_conv_idx\" ON public.\"AiAgentAction\" USING btree (\"conversationId\")",
			"CREATE INDEX \"AiAgentAction_target_idx\" ON public.\"AiAgentAction\" USING btree (\"targetId\")",
		},
	})
}
