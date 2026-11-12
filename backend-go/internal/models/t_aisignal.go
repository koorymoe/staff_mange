// موديل جدول "AiSignal" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiSignal",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "entityType", Type: "text", NotNull: true},
			{Name: "entityId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "payload", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "occurredAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiSignal_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "AiSignal_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AiSignal_pending_idx\" ON public.\"AiSignal\" USING btree (status, \"occurredAt\" DESC)",
			"CREATE UNIQUE INDEX \"AiSignal_unique_event\" ON public.\"AiSignal\" USING btree (kind, \"entityType\", \"entityId\", \"occurredAt\")",
		},
	})
}
