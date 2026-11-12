// موديل جدول "DuplicateCandidate" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DuplicateCandidate",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "entityAId", Type: "text", NotNull: true},
			{Name: "entityBId", Type: "text", NotNull: true},
			{Name: "matchReason", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
			{Name: "detectedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "resolution", Type: "text"},
			{Name: "resolutionNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "DuplicateCandidate_kind_check", Def: "CHECK ((kind = ANY (ARRAY['BOOKING'::text, 'CUSTOMER'::text])))"},
			{Name: "DuplicateCandidate_kind_entityAId_entityBId_key", Def: "UNIQUE (kind, \"entityAId\", \"entityBId\")"},
			{Name: "DuplicateCandidate_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DuplicateCandidate_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "DuplicateCandidate_status_check", Def: "CHECK ((status = ANY (ARRAY['PENDING'::text, 'DISMISSED'::text, 'RESOLVED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"DuplicateCandidate_status_idx\" ON public.\"DuplicateCandidate\" USING btree (kind, status)",
		},
	})
}
