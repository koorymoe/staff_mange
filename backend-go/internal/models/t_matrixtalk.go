// موديل جدول "MatrixTalk" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixTalk",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true, Default: "'محادثة جديدة'::text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "summary", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "summarizedUpTo", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "MatrixTalk_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "MatrixTalk_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"MatrixTalk_employee\" ON public.\"MatrixTalk\" USING btree (\"employeeId\", \"updatedAt\" DESC)",
		},
	})
}
