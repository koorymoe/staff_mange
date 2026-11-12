// موديل جدول "MatrixProposal" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixProposal",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "rationale", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "evidence", Type: "jsonb"},
			{Name: "payload", Type: "jsonb"},
			{Name: "signature", Type: "text", NotNull: true},
			{Name: "source", Type: "text", NotNull: true, Default: "'RULES'::text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "note", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "verdict", Type: "text"},
			{Name: "verifiedAt", Type: "timestamp with time zone"},
			{Name: "leftAtEnd", Type: "integer"},
		},
		Constraints: []Constraint{
			{Name: "MatrixProposal_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "MatrixProposal_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MatrixProposal_verdict_check", Def: "CHECK ((verdict = ANY (ARRAY['CORRECT'::text, 'WRONG'::text])))"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"MatrixProposal_pending_sig\" ON public.\"MatrixProposal\" USING btree (signature) WHERE (status = 'PENDING'::text)",
			"CREATE INDEX \"MatrixProposal_status_idx\" ON public.\"MatrixProposal\" USING btree (status, \"createdAt\" DESC)",
		},
	})
}
