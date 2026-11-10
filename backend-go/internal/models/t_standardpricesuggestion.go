// موديل جدول "StandardPriceSuggestion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "StandardPriceSuggestion",
		Columns: []Column{
			{Name: "key", Type: "text", NotNull: true},
			{Name: "product", Type: "text", NotNull: true},
			{Name: "current", Type: "double precision", NotNull: true},
			{Name: "suggested", Type: "double precision", NotNull: true},
			{Name: "basis", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "StandardPriceSuggestion_pkey", Def: "PRIMARY KEY (key)"},
		},
	})
}
