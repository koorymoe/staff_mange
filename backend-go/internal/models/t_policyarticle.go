// موديل جدول "PolicyArticle" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PolicyArticle",
		Columns: []Column{
			{Name: "no", Type: "integer", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "PolicyArticle_pkey", Def: "PRIMARY KEY (no)"},
		},
	})
}
