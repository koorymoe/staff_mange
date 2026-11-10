// موديل جدول "MatrixTalkMessage" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixTalkMessage",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "chatId", Type: "text", NotNull: true},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "steps", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "image", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "MatrixTalkMessage_chatId_fkey", Def: "FOREIGN KEY (\"chatId\") REFERENCES \"MatrixTalk\"(id) ON DELETE CASCADE"},
			{Name: "MatrixTalkMessage_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MatrixTalkMessage_role_check", Def: "CHECK ((role = ANY (ARRAY['USER'::text, 'ASSISTANT'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"MatrixTalkMessage_chat\" ON public.\"MatrixTalkMessage\" USING btree (\"chatId\", \"createdAt\")",
		},
	})
}
