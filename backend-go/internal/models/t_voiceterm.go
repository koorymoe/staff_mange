// موديل جدول "VoiceTerm" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VoiceTerm",
		Columns: []Column{
			{Name: "term", Type: "text", NotNull: true},
			{Name: "meaning", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "category", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "source", Type: "text", NotNull: true, Default: "'SEED'::text"},
			{Name: "uses", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "VoiceTerm_pkey", Def: "PRIMARY KEY (term)"},
			{Name: "VoiceTerm_source_check", Def: "CHECK ((source = ANY (ARRAY['SEED'::text, 'SYSTEM'::text, 'LEARNED'::text, 'OWNER'::text])))"},
		},
	})
}
