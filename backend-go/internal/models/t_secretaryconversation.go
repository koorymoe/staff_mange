// موديل جدول "SecretaryConversation" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SecretaryConversation",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true, Default: "'محادثة جديدة'::text"},
			{Name: "summary", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "summarizedUpTo", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SecretaryConversation_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "SecretaryConversation_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SecretaryConversation_emp_idx\" ON public.\"SecretaryConversation\" USING btree (\"employeeId\", \"updatedAt\" DESC)",
		},
	})
}
