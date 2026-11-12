// موديل جدول "MatrixForecastOutcome" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixForecastOutcome",
		Columns: []Column{
			{Name: "month", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "forecast", Type: "double precision", NotNull: true},
			{Name: "actual", Type: "double precision", NotNull: true},
			{Name: "errorPct", Type: "double precision", NotNull: true},
			{Name: "verdict", Type: "text", NotNull: true},
			{Name: "snapshotDay", Type: "date", NotNull: true},
			{Name: "evaluatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "MatrixForecastOutcome_pkey", Def: "PRIMARY KEY (month, kind)"},
		},
	})
}
