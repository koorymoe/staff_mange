// موديل جدول "VehicleDailyRating" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleDailyRating",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "ratedDate", Type: "date", NotNull: true, Default: "CURRENT_DATE"},
			{Name: "wash", Type: "integer"},
			{Name: "exteriorClean", Type: "integer"},
			{Name: "exteriorCondition", Type: "integer"},
			{Name: "tireCondition", Type: "integer"},
			{Name: "glassClean", Type: "integer"},
			{Name: "lightsCondition", Type: "integer"},
			{Name: "technicalFaults", Type: "integer"},
			{Name: "faultDescription", Type: "text"},
			{Name: "interiorClean", Type: "integer"},
			{Name: "seatsCondition", Type: "integer"},
			{Name: "interiorDirt", Type: "integer"},
			{Name: "smell", Type: "integer"},
			{Name: "notes", Type: "text"},
			{Name: "recordedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehicleDailyRating_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleDailyRating_recordedById_fkey", Def: "FOREIGN KEY (\"recordedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "VehicleDailyRating_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleDailyRating_vehicleId_idx\" ON public.\"VehicleDailyRating\" USING btree (\"vehicleId\", \"ratedDate\" DESC)",
		},
	})
}
