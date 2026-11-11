// موديل جدول "TrainingProgramParticipant" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingProgramParticipant",
		Columns: []Column{
			{Name: "programId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "passed", Type: "boolean"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "TrainingProgramParticipant_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "TrainingProgramParticipant_pkey", Def: "PRIMARY KEY (\"programId\", \"employeeId\")"},
			{Name: "TrainingProgramParticipant_programId_fkey", Def: "FOREIGN KEY (\"programId\") REFERENCES \"TrainingProgram\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"TrainingProgramParticipant_employee_idx\" ON public.\"TrainingProgramParticipant\" USING btree (\"employeeId\")",
		},
	})
}
