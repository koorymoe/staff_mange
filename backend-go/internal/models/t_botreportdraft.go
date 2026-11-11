// موديل جدول "BotReportDraft" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BotReportDraft",
		Columns: []Column{
			{Name: "chatId", Type: "bigint", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "body", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "photos", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "BotReportDraft_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "BotReportDraft_pkey", Def: "PRIMARY KEY (\"chatId\")"},
		},
	})
}
