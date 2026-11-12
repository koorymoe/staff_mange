// موديل جدول "DesignForm" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DesignForm",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "publicToken", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "DesignForm_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DesignForm_publicToken_key", Def: "UNIQUE (\"publicToken\")"},
		},
	})
}
