// موديل جدول "SimMastery" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimMastery",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "exerciseId", Type: "text", NotNull: true},
			{Name: "bestScore", Type: "integer", NotNull: true, Default: "0"},
			{Name: "attempts", Type: "integer", NotNull: true, Default: "0"},
			{Name: "passed", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "firstPassAt", Type: "timestamp with time zone"},
			{Name: "lastAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimMastery_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "SimMastery_exerciseId_fkey", Def: "FOREIGN KEY (\"exerciseId\") REFERENCES \"SimExercise\"(id) ON DELETE CASCADE"},
			{Name: "SimMastery_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"SimMastery_uniq\" ON public.\"SimMastery\" USING btree (\"employeeId\", \"exerciseId\")",
		},
	})
}
