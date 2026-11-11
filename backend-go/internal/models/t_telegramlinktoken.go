// موديل جدول "TelegramLinkToken" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TelegramLinkToken",
		Columns: []Column{
			{Name: "token", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "expiresAt", Type: "timestamp without time zone", NotNull: true},
			{Name: "usedAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "TelegramLinkToken_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "TelegramLinkToken_pkey", Def: "PRIMARY KEY (token)"},
		},
	})
}
