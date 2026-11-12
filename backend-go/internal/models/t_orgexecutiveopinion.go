// موديل جدول "OrgExecutiveOpinion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgExecutiveOpinion",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "period", Type: "text", NotNull: true},
			{Name: "rating", Type: "integer", NotNull: true},
			{Name: "strengths", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "weaknesses", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "recommendation", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "facts", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "requestedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "OrgExecutiveOpinion_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "OrgExecutiveOpinion_employeeId_period_key", Def: "UNIQUE (\"employeeId\", period)"},
			{Name: "OrgExecutiveOpinion_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "OrgExecutiveOpinion_rating_check", Def: "CHECK (((rating >= 1) AND (rating <= 5)))"},
			{Name: "OrgExecutiveOpinion_requestedById_fkey", Def: "FOREIGN KEY (\"requestedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
