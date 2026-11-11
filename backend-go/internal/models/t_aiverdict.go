// موديل جدول "AiVerdict" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiVerdict",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "signalId", Type: "text", NotNull: true},
			{Name: "source", Type: "text", NotNull: true, Default: "'RULES'::text"},
			{Name: "modelName", Type: "text"},
			{Name: "headline", Type: "text", NotNull: true},
			{Name: "reasoning", Type: "text"},
			{Name: "confidence", Type: "integer", NotNull: true, Default: "0"},
			{Name: "severity", Type: "text", NotNull: true, Default: "'INFO'::text"},
			{Name: "blameEmployeeId", Type: "text"},
			{Name: "suggestion", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiVerdict_blameEmployeeId_fkey", Def: "FOREIGN KEY (\"blameEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "AiVerdict_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "AiVerdict_signalId_fkey", Def: "FOREIGN KEY (\"signalId\") REFERENCES \"AiSignal\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"AiVerdict_severity_idx\" ON public.\"AiVerdict\" USING btree (severity, \"createdAt\" DESC)",
			"CREATE INDEX \"AiVerdict_signal_idx\" ON public.\"AiVerdict\" USING btree (\"signalId\")",
		},
	})
}
