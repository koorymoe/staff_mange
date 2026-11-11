// موديل جدول "IraqiPhrase" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "IraqiPhrase",
		Columns: []Column{
			{Name: "phrase", Type: "text", NotNull: true},
			{Name: "meaning", Type: "text", NotNull: true},
			{Name: "uses", Type: "integer", NotNull: true, Default: "1"},
			{Name: "firstSeen", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "lastSeen", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "IraqiPhrase_pkey", Def: "PRIMARY KEY (phrase)"},
		},
	})
}
