// موديل جدول "TrainingNeed" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingNeed",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "issueId", Type: "text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "verdict", Type: "text", NotNull: true},
			{Name: "areas", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "rules", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "evidence", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "baselinePct", Type: "double precision"},
			{Name: "followUpAt", Type: "timestamp with time zone", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "currentPct", Type: "double precision"},
			{Name: "resultNote", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "TrainingNeed_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "TrainingNeed_issueId_employeeId_key", Def: "UNIQUE (\"issueId\", \"employeeId\")"},
			{Name: "TrainingNeed_issueId_fkey", Def: "FOREIGN KEY (\"issueId\") REFERENCES \"WorkplaceIssue\"(id) ON DELETE CASCADE"},
			{Name: "TrainingNeed_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"TrainingNeed_due_idx\" ON public.\"TrainingNeed\" USING btree (status, \"followUpAt\")",
		},
	})
}
