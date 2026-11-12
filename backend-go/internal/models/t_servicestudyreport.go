// موديل جدول "ServiceStudyReport" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ServiceStudyReport",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "serviceStudyId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "content", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ServiceStudyReport_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "ServiceStudyReport_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ServiceStudyReport_serviceStudyId_fkey", Def: "FOREIGN KEY (\"serviceStudyId\") REFERENCES \"ServiceStudy\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ServiceStudyReport_serviceStudyId_idx\" ON public.\"ServiceStudyReport\" USING btree (\"serviceStudyId\")",
		},
	})
}
