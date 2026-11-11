// موديل جدول "WaMessage" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "WaMessage",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "conversationId", Type: "text", NotNull: true},
			{Name: "direction", Type: "text", NotNull: true},
			{Name: "author", Type: "text", NotNull: true},
			{Name: "authorId", Type: "text"},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "metaId", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "WaMessage_authorId_fkey", Def: "FOREIGN KEY (\"authorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "WaMessage_author_check", Def: "CHECK ((author = ANY (ARRAY['CUSTOMER'::text, 'AGENT'::text, 'HUMAN'::text])))"},
			{Name: "WaMessage_conversationId_fkey", Def: "FOREIGN KEY (\"conversationId\") REFERENCES \"WaConversation\"(id) ON DELETE CASCADE"},
			{Name: "WaMessage_direction_check", Def: "CHECK ((direction = ANY (ARRAY['IN'::text, 'OUT'::text])))"},
			{Name: "WaMessage_metaId_key", Def: "UNIQUE (\"metaId\")"},
			{Name: "WaMessage_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"WaMessage_conv_idx\" ON public.\"WaMessage\" USING btree (\"conversationId\", \"createdAt\")",
		},
	})
}
