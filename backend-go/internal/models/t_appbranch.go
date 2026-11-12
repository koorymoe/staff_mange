// موديل جدول "AppBranch" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AppBranch",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "address", Type: "text"},
			{Name: "phone", Type: "text"},
			{Name: "mapUrl", Type: "text"},
			{Name: "ord", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "AppBranch_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
