// موديل جدول "MissionEvent" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MissionEvent",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "missionId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "action", Type: "text", NotNull: true},
			{Name: "lat", Type: "double precision"},
			{Name: "lng", Type: "double precision"},
			{Name: "note", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "MissionEvent_missionId_fkey", Def: "FOREIGN KEY (\"missionId\") REFERENCES \"Mission\"(id) ON DELETE CASCADE"},
			{Name: "MissionEvent_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"MissionEvent_missionId_idx\" ON public.\"MissionEvent\" USING btree (\"missionId\")",
		},
	})
}
