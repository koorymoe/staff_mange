// موديل جدول "MatrixSuggestion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixSuggestion",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "suggested", Type: "jsonb", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "outcome", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "actual", Type: "jsonb"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "MatrixSuggestion_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "MatrixSuggestion_bookingId_kind_key", Def: "UNIQUE (\"bookingId\", kind)"},
			{Name: "MatrixSuggestion_kind_check", Def: "CHECK ((kind = ANY (ARRAY['SCHEDULE'::text, 'CREW'::text])))"},
			{Name: "MatrixSuggestion_outcome_check", Def: "CHECK ((outcome = ANY (ARRAY['PENDING'::text, 'ACCEPTED'::text, 'PARTIAL'::text, 'CHANGED'::text])))"},
			{Name: "MatrixSuggestion_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"MatrixSuggestion_outcome_idx\" ON public.\"MatrixSuggestion\" USING btree (outcome, \"createdAt\")",
		},
	})
}
