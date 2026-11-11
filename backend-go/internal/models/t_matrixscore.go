// موديل جدول "MatrixScore" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixScore",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "source", Type: "text", NotNull: true},
			{Name: "sourceId", Type: "text", NotNull: true},
			{Name: "sourceLabel", Type: "text"},
			{Name: "rule", Type: "text", NotNull: true},
			{Name: "points", Type: "integer", NotNull: true},
			{Name: "maxPoints", Type: "integer", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "at", Type: "timestamp with time zone", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "cancelledAt", Type: "timestamp with time zone"},
			{Name: "cancelledById", Type: "text"},
			{Name: "cancelNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "MatrixScore_cancelledById_fkey", Def: "FOREIGN KEY (\"cancelledById\") REFERENCES \"Employee\"(id)"},
			{Name: "MatrixScore_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "MatrixScore_employeeId_source_sourceId_rule_key", Def: "UNIQUE (\"employeeId\", source, \"sourceId\", rule)"},
			{Name: "MatrixScore_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"MatrixScore_emp_at_idx\" ON public.\"MatrixScore\" USING btree (\"employeeId\", at)",
		},
	})
}
