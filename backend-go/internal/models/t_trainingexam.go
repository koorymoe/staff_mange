// موديل جدول "TrainingExam" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingExam",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "month", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "questionIds", Type: "text[]", NotNull: true},
			{Name: "minutes", Type: "integer", NotNull: true, Default: "15"},
			{Name: "passPct", Type: "integer", NotNull: true, Default: "60"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "TrainingExam_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "TrainingExam_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "TrainingExam_role_month_key", Def: "UNIQUE (role, month)"},
		},
	})
}
