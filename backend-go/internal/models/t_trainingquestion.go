// موديل جدول "TrainingQuestion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingQuestion",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "curriculumId", Type: "text"},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "options", Type: "text[]", NotNull: true},
			{Name: "correct", Type: "integer", NotNull: true},
			{Name: "approved", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "source", Type: "text", NotNull: true, Default: "'MANUAL'::text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "TrainingQuestion_curriculumId_fkey", Def: "FOREIGN KEY (\"curriculumId\") REFERENCES \"TrainingCurriculum\"(id) ON DELETE SET NULL"},
			{Name: "TrainingQuestion_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
