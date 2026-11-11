// موديل جدول "OrgEvaluation" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgEvaluation",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "period", Type: "text", NotNull: true},
			{Name: "score", Type: "numeric"},
			{Name: "components", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "decision", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "OrgEvaluation_decision_check", Def: "CHECK ((decision = ANY (ARRAY['OK'::text, 'WARN'::text, 'REVOKE'::text, 'NODATA'::text])))"},
			{Name: "OrgEvaluation_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "OrgEvaluation_pkey", Def: "PRIMARY KEY (\"employeeId\", period)"},
		},
	})
}
