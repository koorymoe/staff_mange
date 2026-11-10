// موديل جدول "MatrixForecastLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixForecastLog",
		Columns: []Column{
			{Name: "day", Type: "date", NotNull: true},
			{Name: "month", Type: "text", NotNull: true},
			{Name: "expected", Type: "double precision", NotNull: true},
			{Name: "pipeline", Type: "double precision", NotNull: true},
			{Name: "pace", Type: "double precision"},
			{Name: "mtd", Type: "double precision", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "expectedJobs", Type: "integer"},
			{Name: "mtdJobs", Type: "integer"},
		},
		Constraints: []Constraint{
			{Name: "MatrixForecastLog_pkey", Def: "PRIMARY KEY (day)"},
		},
		Indexes: []string{
			"CREATE INDEX \"MatrixForecastLog_month_idx\" ON public.\"MatrixForecastLog\" USING btree (month)",
		},
	})
}
