// موديل جدول "AgentFeedback" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AgentFeedback",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "agent", Type: "text", NotNull: true},
			{Name: "messageId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "good", Type: "boolean", NotNull: true},
			{Name: "note", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AgentFeedback_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "AgentFeedback_messageId_employeeId_key", Def: "UNIQUE (\"messageId\", \"employeeId\")"},
			{Name: "AgentFeedback_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
