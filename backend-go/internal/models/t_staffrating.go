// موديل جدول "StaffRating" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "StaffRating",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "raterId", Type: "text", NotNull: true},
			{Name: "rateeId", Type: "text", NotNull: true},
			{Name: "stage", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "period", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "score", Type: "integer", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "StaffRating_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "StaffRating_rateeId_fkey", Def: "FOREIGN KEY (\"rateeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "StaffRating_raterId_fkey", Def: "FOREIGN KEY (\"raterId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "StaffRating_raterId_rateeId_stage_bookingId_period_key", Def: "UNIQUE (\"raterId\", \"rateeId\", stage, \"bookingId\", period)"},
			{Name: "StaffRating_score_check", Def: "CHECK (((score >= 1) AND (score <= 5)))"},
		},
		Indexes: []string{
			"CREATE INDEX \"StaffRating_ratee_idx\" ON public.\"StaffRating\" USING btree (\"rateeId\", \"createdAt\")",
		},
	})
}
