// موديل جدول "AuditTrail" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AuditTrail",
		Columns: []Column{
			{Name: "id", Type: "bigint", NotNull: true, Default: "nextval('\"AuditTrail_id_seq\"'::regclass)"},
			{Name: "at", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "entity", Type: "text", NotNull: true},
			{Name: "entityId", Type: "text"},
			{Name: "action", Type: "text", NotNull: true},
			{Name: "actorId", Type: "text"},
			{Name: "before", Type: "jsonb"},
			{Name: "after", Type: "jsonb"},
		},
		Constraints: []Constraint{
			{Name: "AuditTrail_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AuditTrail_at_idx\" ON public.\"AuditTrail\" USING btree (at DESC)",
			"CREATE INDEX \"AuditTrail_entity_idx\" ON public.\"AuditTrail\" USING btree (entity, \"entityId\", at DESC)",
		},
	})
}
