// موديل جدول "KpiEvaluation" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "KpiEvaluation",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "evaluatorId", Type: "text", NotNull: true},
			{Name: "points", Type: "integer", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "deductionAmount", Type: "double precision", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "cancelled", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "cancelledAt", Type: "timestamp without time zone"},
			{Name: "cancelledByEmployeeId", Type: "text"},
			{Name: "cancelNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "KpiEvaluation_cancelledByEmployeeId_fkey", Def: "FOREIGN KEY (\"cancelledByEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "KpiEvaluation_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "KpiEvaluation_evaluatorId_fkey", Def: "FOREIGN KEY (\"evaluatorId\") REFERENCES \"Employee\"(id)"},
			{Name: "KpiEvaluation_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"KpiEvaluation_employeeId_idx\" ON public.\"KpiEvaluation\" USING btree (\"employeeId\")",
		},
	})
}
