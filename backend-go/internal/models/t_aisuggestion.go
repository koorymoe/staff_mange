// موديل جدول "AiSuggestion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiSuggestion",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "agent", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "details", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "needsCode", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "status", Type: "text", NotNull: true, Default: "'NEW'::text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "verifyResult", Type: "text"},
			{Name: "verifyNote", Type: "text"},
			{Name: "verifiedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "AiSuggestion_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "AiSuggestion_status_check", Def: "CHECK ((status = ANY (ARRAY['NEW'::text, 'ACCEPTED'::text, 'REJECTED'::text, 'DONE'::text, 'VERIFYING'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"AiSuggestion_status_idx\" ON public.\"AiSuggestion\" USING btree (status, \"createdAt\" DESC)",
		},
	})
}
