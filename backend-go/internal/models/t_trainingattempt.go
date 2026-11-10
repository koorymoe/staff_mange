// موديل جدول "TrainingAttempt" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingAttempt",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "examId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "answers", Type: "integer[]"},
			{Name: "score", Type: "integer"},
			{Name: "total", Type: "integer"},
			{Name: "startedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "finishedAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "TrainingAttempt_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "TrainingAttempt_examId_employeeId_key", Def: "UNIQUE (\"examId\", \"employeeId\")"},
			{Name: "TrainingAttempt_examId_fkey", Def: "FOREIGN KEY (\"examId\") REFERENCES \"TrainingExam\"(id) ON DELETE CASCADE"},
			{Name: "TrainingAttempt_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
