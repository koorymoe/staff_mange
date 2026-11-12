// موديل جدول "SecretaryAudio" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SecretaryAudio",
		Columns: []Column{
			{Name: "messageId", Type: "text", NotNull: true},
			{Name: "contentType", Type: "text", NotNull: true},
			{Name: "data", Type: "bytea", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SecretaryAudio_messageId_fkey", Def: "FOREIGN KEY (\"messageId\") REFERENCES \"SecretaryMessage\"(id) ON DELETE CASCADE"},
			{Name: "SecretaryAudio_pkey", Def: "PRIMARY KEY (\"messageId\")"},
		},
	})
}
