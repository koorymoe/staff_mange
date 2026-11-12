// موديل جدول "AiWorkWindow" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiWorkWindow",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "startHour", Type: "integer", NotNull: true, Default: "9"},
			{Name: "endHour", Type: "integer", NotNull: true, Default: "24"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiWorkWindow_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
