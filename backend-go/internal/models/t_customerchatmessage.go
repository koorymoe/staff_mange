// موديل جدول "CustomerChatMessage" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CustomerChatMessage",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "accountId", Type: "text", NotNull: true},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "content", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "CustomerChatMessage_accountId_fkey", Def: "FOREIGN KEY (\"accountId\") REFERENCES \"CustomerAccount\"(id) ON DELETE CASCADE"},
			{Name: "CustomerChatMessage_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX customer_chat_acc ON public.\"CustomerChatMessage\" USING btree (\"accountId\", \"createdAt\")",
		},
	})
}
