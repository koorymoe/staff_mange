// موديل جدول "VehicleMissionRating" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleMissionRating",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "missionId", Type: "text", NotNull: true},
			{Name: "ratedById", Type: "text", NotNull: true},
			{Name: "commitment", Type: "integer", NotNull: true},
			{Name: "vehicleCare", Type: "integer", NotNull: true},
			{Name: "driving", Type: "integer", NotNull: true},
			{Name: "cleanliness", Type: "integer", NotNull: true},
			{Name: "notes", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehicleMissionRating_missionId_fkey", Def: "FOREIGN KEY (\"missionId\") REFERENCES \"VehicleMission\"(id) ON DELETE CASCADE"},
			{Name: "VehicleMissionRating_missionId_key", Def: "UNIQUE (\"missionId\")"},
			{Name: "VehicleMissionRating_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleMissionRating_ratedById_fkey", Def: "FOREIGN KEY (\"ratedById\") REFERENCES \"Employee\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleMissionRating_missionId_idx\" ON public.\"VehicleMissionRating\" USING btree (\"missionId\")",
		},
	})
}
