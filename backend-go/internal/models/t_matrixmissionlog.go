// موديل جدول "MatrixMissionLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixMissionLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "missionId", Type: "text", NotNull: true},
			{Name: "day", Type: "date", NotNull: true},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "MatrixMissionLog_missionId_day_key", Def: "UNIQUE (\"missionId\", day)"},
			{Name: "MatrixMissionLog_missionId_fkey", Def: "FOREIGN KEY (\"missionId\") REFERENCES \"MatrixMission\"(id) ON DELETE CASCADE"},
			{Name: "MatrixMissionLog_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
