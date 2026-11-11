// موديل جدول "TrainingMentor" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingMentor",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "mentorId", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "TrainingMentor_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "TrainingMentor_mentorId_fkey", Def: "FOREIGN KEY (\"mentorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "TrainingMentor_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
		},
	})
}
