// موديل جدول "AchievementVoice" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AchievementVoice",
		Columns: []Column{
			{Name: "achievementId", Type: "text", NotNull: true},
			{Name: "fileKey", Type: "text", NotNull: true},
			{Name: "contentType", Type: "text", NotNull: true},
			{Name: "seconds", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "analysisStatus", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "attempts", Type: "integer", NotNull: true, Default: "0"},
			{Name: "transcript", Type: "text"},
			{Name: "summary", Type: "text"},
			{Name: "tasks", Type: "jsonb"},
			{Name: "style", Type: "text"},
			{Name: "analysisError", Type: "text"},
			{Name: "analyzedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "AchievementVoice_achievementId_fkey", Def: "FOREIGN KEY (\"achievementId\") REFERENCES \"Achievement\"(id) ON DELETE CASCADE"},
			{Name: "AchievementVoice_pkey", Def: "PRIMARY KEY (\"achievementId\")"},
		},
	})
}
