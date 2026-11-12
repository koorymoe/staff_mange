// موديل جدول "FreeWorkReason" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "FreeWorkReason",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "label", Type: "text", NotNull: true},
			{Name: "sortOrder", Type: "integer", NotNull: true, Default: "0"},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "needsNote", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "FreeWorkReason_label_key", Def: "UNIQUE (label)"},
			{Name: "FreeWorkReason_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
