// موديل جدول "VehicleMonthlyStatus" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleMonthlyStatus",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "vehicleId", Type: "text", NotNull: true},
			{Name: "month", Type: "text", NotNull: true},
			{Name: "hasIssue", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "issueDescription", Type: "text"},
			{Name: "resolved", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "notes", Type: "text"},
			{Name: "recordedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehicleMonthlyStatus_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "VehicleMonthlyStatus_recordedById_fkey", Def: "FOREIGN KEY (\"recordedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "VehicleMonthlyStatus_vehicleId_fkey", Def: "FOREIGN KEY (\"vehicleId\") REFERENCES \"Vehicle\"(id) ON DELETE CASCADE"},
			{Name: "VehicleMonthlyStatus_vehicleId_month_key", Def: "UNIQUE (\"vehicleId\", month)"},
		},
	})
}
