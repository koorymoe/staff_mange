// موديل جدول "LeaderSkillRating" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "LeaderSkillRating",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "skill", Type: "text", NotNull: true},
			{Name: "score", Type: "integer", NotNull: true},
			{Name: "ratedById", Type: "text"},
			{Name: "ratedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "LeaderSkillRating_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "LeaderSkillRating_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "LeaderSkillRating_ratedById_fkey", Def: "FOREIGN KEY (\"ratedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "LeaderSkillRating_score_check", Def: "CHECK (((score >= 1) AND (score <= 10)))"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"LeaderSkillRating_employee_skill_key\" ON public.\"LeaderSkillRating\" USING btree (\"employeeId\", skill)",
		},
	})
}
