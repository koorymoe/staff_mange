// موديل جدول "DesignSection" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DesignSection",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "setId", Type: "text", NotNull: true},
			{Name: "discipline", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'TODO'::text"},
			{Name: "assigneeId", Type: "text"},
			{Name: "dueDate", Type: "date"},
			{Name: "taskId", Type: "text"},
			{Name: "submittedAt", Type: "timestamp with time zone"},
			{Name: "reviewNote", Type: "text"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "DesignSection_assigneeId_fkey", Def: "FOREIGN KEY (\"assigneeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "DesignSection_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DesignSection_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "DesignSection_setId_discipline_key", Def: "UNIQUE (\"setId\", discipline)"},
			{Name: "DesignSection_setId_fkey", Def: "FOREIGN KEY (\"setId\") REFERENCES \"DesignSet\"(id) ON DELETE CASCADE"},
			{Name: "DesignSection_status_check", Def: "CHECK ((status = ANY (ARRAY['TODO'::text, 'IN_PROGRESS'::text, 'REVIEW'::text, 'APPROVED'::text, 'RETURNED'::text])))"},
			{Name: "DesignSection_taskId_fkey", Def: "FOREIGN KEY (\"taskId\") REFERENCES \"ExtraTask\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"DesignSection_assignee_idx\" ON public.\"DesignSection\" USING btree (\"assigneeId\")",
		},
	})
}
