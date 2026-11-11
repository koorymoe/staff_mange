// موديل جدول "PeerCheckin" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PeerCheckin",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "week", Type: "date", NotNull: true},
			{Name: "mood", Type: "integer"},
			{Name: "worry", Type: "text"},
			{Name: "suggestion", Type: "text"},
			{Name: "nothing", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "urgent", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "PeerCheckin_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "PeerCheckin_employeeId_week_key", Def: "UNIQUE (\"employeeId\", week)"},
			{Name: "PeerCheckin_mood_check", Def: "CHECK (((mood >= 1) AND (mood <= 5)))"},
			{Name: "PeerCheckin_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
