// موديل جدول "TrainingCurriculum" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingCurriculum",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "details", Type: "text"},
			{Name: "ord", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "TrainingCurriculum_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
