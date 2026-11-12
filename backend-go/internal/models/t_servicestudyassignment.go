// موديل جدول "ServiceStudyAssignment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ServiceStudyAssignment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "serviceStudyId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ServiceStudyAssignment_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "ServiceStudyAssignment_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ServiceStudyAssignment_serviceStudyId_employeeId_key", Def: "UNIQUE (\"serviceStudyId\", \"employeeId\")"},
			{Name: "ServiceStudyAssignment_serviceStudyId_fkey", Def: "FOREIGN KEY (\"serviceStudyId\") REFERENCES \"ServiceStudy\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ServiceStudyAssignment_serviceStudyId_idx\" ON public.\"ServiceStudyAssignment\" USING btree (\"serviceStudyId\")",
		},
	})
}
