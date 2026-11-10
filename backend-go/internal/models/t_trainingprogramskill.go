// موديل جدول "TrainingProgramSkill" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingProgramSkill",
		Columns: []Column{
			{Name: "programId", Type: "text", NotNull: true},
			{Name: "skillId", Type: "text", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "TrainingProgramSkill_pkey", Def: "PRIMARY KEY (\"programId\", \"skillId\")"},
			{Name: "TrainingProgramSkill_programId_fkey", Def: "FOREIGN KEY (\"programId\") REFERENCES \"TrainingProgram\"(id) ON DELETE CASCADE"},
			{Name: "TrainingProgramSkill_skillId_fkey", Def: "FOREIGN KEY (\"skillId\") REFERENCES \"Skill\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"TrainingProgramSkill_skill_idx\" ON public.\"TrainingProgramSkill\" USING btree (\"skillId\")",
		},
	})
}
