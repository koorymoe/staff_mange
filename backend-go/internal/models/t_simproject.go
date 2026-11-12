// موديل جدول "SimProject" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimProject",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "domain", Type: "text", NotNull: true},
			{Name: "doc", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "notes", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimProject_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "SimProject_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SimProject_emp_idx\" ON public.\"SimProject\" USING btree (\"employeeId\", \"updatedAt\" DESC)",
		},
	})
}
