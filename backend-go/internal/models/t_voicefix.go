// موديل جدول "VoiceFix" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VoiceFix",
		Columns: []Column{
			{Name: "wrong", Type: "text", NotNull: true},
			{Name: "right", Type: "text", NotNull: true},
			{Name: "source", Type: "text", NotNull: true, Default: "'SEED'::text"},
			{Name: "uses", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "VoiceFix_pkey", Def: "PRIMARY KEY (wrong)"},
		},
	})
}
