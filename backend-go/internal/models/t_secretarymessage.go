// موديل جدول "SecretaryMessage" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SecretaryMessage",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "text", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "transcript", Type: "text"},
			{Name: "cards", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "actionId", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "conversationId", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "SecretaryMessage_conversationId_fkey", Def: "FOREIGN KEY (\"conversationId\") REFERENCES \"SecretaryConversation\"(id) ON DELETE CASCADE"},
			{Name: "SecretaryMessage_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "SecretaryMessage_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SecretaryMessage_role_check", Def: "CHECK ((role = ANY (ARRAY['user'::text, 'assistant'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"SecretaryMessage_conv_idx\" ON public.\"SecretaryMessage\" USING btree (\"conversationId\", \"createdAt\")",
			"CREATE INDEX \"SecretaryMessage_emp_idx\" ON public.\"SecretaryMessage\" USING btree (\"employeeId\", \"createdAt\" DESC)",
		},
	})
}
