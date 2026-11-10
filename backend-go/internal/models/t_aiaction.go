// موديل جدول "AiAction" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiAction",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "entityType", Type: "text", NotNull: true},
			{Name: "entityId", Type: "text", NotNull: true},
			{Name: "period", Type: "text", NotNull: true},
			{Name: "targetEmployeeId", Type: "text"},
			{Name: "targetLabel", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "summary", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'DONE'::text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "undoneById", Type: "text"},
			{Name: "undoneAt", Type: "timestamp with time zone"},
			{Name: "details", Type: "jsonb"},
			{Name: "resolvedAt", Type: "timestamp with time zone"},
			{Name: "escalatedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "AiAction_kind_entityId_period_key", Def: "UNIQUE (kind, \"entityId\", period)"},
			{Name: "AiAction_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "AiAction_targetEmployeeId_fkey", Def: "FOREIGN KEY (\"targetEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "AiAction_undoneById_fkey", Def: "FOREIGN KEY (\"undoneById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"AiAction_createdAt_idx\" ON public.\"AiAction\" USING btree (\"createdAt\" DESC)",
		},
	})
}
