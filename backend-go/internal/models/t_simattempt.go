// موديل جدول "SimAttempt" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimAttempt",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "exerciseId", Type: "text", NotNull: true},
			{Name: "exerciseVersion", Type: "integer", NotNull: true, Default: "1"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'IN_PROGRESS'::text"},
			{Name: "score", Type: "integer"},
			{Name: "stepsTotal", Type: "integer", NotNull: true, Default: "0"},
			{Name: "stepsPassed", Type: "integer", NotNull: true, Default: "0"},
			{Name: "hintsUsed", Type: "integer", NotNull: true, Default: "0"},
			{Name: "wrongCount", Type: "integer", NotNull: true, Default: "0"},
			{Name: "durationSec", Type: "integer"},
			{Name: "state", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "startedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "finishedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimAttempt_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "SimAttempt_exerciseId_fkey", Def: "FOREIGN KEY (\"exerciseId\") REFERENCES \"SimExercise\"(id) ON DELETE CASCADE"},
			{Name: "SimAttempt_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SimAttempt_emp_idx\" ON public.\"SimAttempt\" USING btree (\"employeeId\", \"startedAt\" DESC)",
			"CREATE INDEX \"SimAttempt_ex_idx\" ON public.\"SimAttempt\" USING btree (\"exerciseId\", status)",
		},
	})
}
