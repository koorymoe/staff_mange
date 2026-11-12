// موديل جدول "EmployeeCharacter" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeCharacter",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "persona", Type: "text"},
			{Name: "calmKey", Type: "text"},
			{Name: "happyKey", Type: "text"},
			{Name: "angryKey", Type: "text"},
			{Name: "prompt", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "error", Type: "text"},
			{Name: "generatedAt", Type: "timestamp without time zone"},
			{Name: "generatedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeCharacter_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeeCharacter_employeeId_key", Def: "UNIQUE (\"employeeId\")"},
			{Name: "EmployeeCharacter_generatedById_fkey", Def: "FOREIGN KEY (\"generatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "EmployeeCharacter_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"EmployeeCharacter_status_idx\" ON public.\"EmployeeCharacter\" USING btree (status)",
		},
	})
}
