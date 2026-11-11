// موديل جدول "AiEvidence" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiEvidence",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "signalId", Type: "text", NotNull: true},
			{Name: "facts", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "gaps", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "collectedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiEvidence_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "AiEvidence_signalId_fkey", Def: "FOREIGN KEY (\"signalId\") REFERENCES \"AiSignal\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"AiEvidence_signal_unique\" ON public.\"AiEvidence\" USING btree (\"signalId\")",
		},
	})
}
