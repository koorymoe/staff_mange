// موديل جدول "SystemSwitch" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SystemSwitch",
		Columns: []Column{
			{Name: "key", Type: "text", NotNull: true},
			{Name: "enabled", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "updatedBy", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "SystemSwitch_pkey", Def: "PRIMARY KEY (key)"},
		},
	})
}
