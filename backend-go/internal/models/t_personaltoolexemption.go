// موديل جدول "PersonalToolExemption" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PersonalToolExemption",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "toolName", Type: "text", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "byEmployeeId", Type: "text"},
			{Name: "byName", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "PersonalToolExemption_byEmployeeId_fkey", Def: "FOREIGN KEY (\"byEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "PersonalToolExemption_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "PersonalToolExemption_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"PersonalToolExemption_emp_tool_idx\" ON public.\"PersonalToolExemption\" USING btree (\"employeeId\", \"toolName\")",
		},
	})
}
