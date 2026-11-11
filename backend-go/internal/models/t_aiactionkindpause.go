// موديل جدول "AiActionKindPause" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiActionKindPause",
		Columns: []Column{
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "pausedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "resumedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "AiActionKindPause_pkey", Def: "PRIMARY KEY (kind)"},
		},
	})
}
