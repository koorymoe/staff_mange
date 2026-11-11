// موديل جدول "StandardPrice" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "StandardPrice",
		Columns: []Column{
			{Name: "key", Type: "text", NotNull: true},
			{Name: "product", Type: "text", NotNull: true},
			{Name: "price", Type: "double precision", NotNull: true},
			{Name: "source", Type: "text", NotNull: true, Default: "'AUTO'::text"},
			{Name: "setById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "StandardPrice_pkey", Def: "PRIMARY KEY (key)"},
		},
	})
}
