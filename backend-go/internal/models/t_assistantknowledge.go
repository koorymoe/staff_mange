// موديل جدول "AssistantKnowledge" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AssistantKnowledge",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "topic", Type: "text", NotNull: true},
			{Name: "content", Type: "text", NotNull: true},
			{Name: "learnedFromEmployeeId", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "AssistantKnowledge_learnedFromEmployeeId_fkey", Def: "FOREIGN KEY (\"learnedFromEmployeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "AssistantKnowledge_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AssistantKnowledge_createdAt_idx\" ON public.\"AssistantKnowledge\" USING btree (\"createdAt\")",
			"CREATE INDEX \"AssistantKnowledge_topic_idx\" ON public.\"AssistantKnowledge\" USING btree (topic)",
		},
	})
}
