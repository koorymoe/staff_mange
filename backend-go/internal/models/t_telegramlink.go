// موديل جدول "TelegramLink" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TelegramLink",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "chatId", Type: "bigint", NotNull: true},
			{Name: "tgUsername", Type: "text"},
			{Name: "linkedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "TelegramLink_chatId_key", Def: "UNIQUE (\"chatId\")"},
			{Name: "TelegramLink_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "TelegramLink_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
		},
	})
}
