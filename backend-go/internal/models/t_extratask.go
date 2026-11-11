// موديل جدول "ExtraTask" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ExtraTask",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "description", Type: "text"},
			{Name: "assignedToId", Type: "text", NotNull: true},
			{Name: "assignedById", Type: "text"},
			{Name: "priority", Type: "text", NotNull: true, Default: "'NORMAL'::text"},
			{Name: "dueAt", Type: "timestamp with time zone"},
			{Name: "status", Type: "text", NotNull: true, Default: "'NEW'::text"},
			{Name: "seenAt", Type: "timestamp with time zone"},
			{Name: "startedAt", Type: "timestamp with time zone"},
			{Name: "doneAt", Type: "timestamp with time zone"},
			{Name: "doneNote", Type: "text"},
			{Name: "cancelledAt", Type: "timestamp with time zone"},
			{Name: "cancelReason", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "clause", Type: "integer"},
		},
		Constraints: []Constraint{
			{Name: "ExtraTask_assignedById_fkey", Def: "FOREIGN KEY (\"assignedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ExtraTask_assignedToId_fkey", Def: "FOREIGN KEY (\"assignedToId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "ExtraTask_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ExtraTask_assignee_idx\" ON public.\"ExtraTask\" USING btree (\"assignedToId\", status, \"createdAt\" DESC)",
			"CREATE INDEX \"ExtraTask_assigner_idx\" ON public.\"ExtraTask\" USING btree (\"assignedById\", \"createdAt\" DESC)",
		},
	})
}
