// موديل جدول "VehicleWashRating" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleWashRating",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "dailyRatingId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "score", Type: "integer", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehicleWashRating_dailyRatingId_fkey", Def: "FOREIGN KEY (\"dailyRatingId\") REFERENCES \"VehicleDailyRating\"(id) ON DELETE CASCADE"},
			{Name: "VehicleWashRating_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "VehicleWashRating_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleWashRating_dailyRatingId_idx\" ON public.\"VehicleWashRating\" USING btree (\"dailyRatingId\")",
			"CREATE INDEX \"VehicleWashRating_employeeId_idx\" ON public.\"VehicleWashRating\" USING btree (\"employeeId\")",
		},
	})
}
