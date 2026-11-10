// موديل جدول "EmployeeSkill" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeSkill",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "skillId", Type: "text", NotNull: true},
			{Name: "canPerform", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeSkill_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeeSkill_employeeId_skillId_key", Def: "UNIQUE (\"employeeId\", \"skillId\")"},
			{Name: "EmployeeSkill_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "EmployeeSkill_skillId_fkey", Def: "FOREIGN KEY (\"skillId\") REFERENCES \"Skill\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"EmployeeSkill_skillId_idx\" ON public.\"EmployeeSkill\" USING btree (\"skillId\")",
		},
	})
}
