// موديل جدول "TrainingPlan" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingPlan",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "month", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "role", Type: "text"},
			{Name: "employeeIds", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "plannedDate", Type: "date"},
			{Name: "trainerId", Type: "text"},
			{Name: "inspectorId", Type: "text"},
			{Name: "inspectionResult", Type: "text"},
			{Name: "inspectionNotes", Type: "text"},
			{Name: "inspectedAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "TrainingPlan_inspectorId_fkey", Def: "FOREIGN KEY (\"inspectorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "TrainingPlan_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "TrainingPlan_trainerId_fkey", Def: "FOREIGN KEY (\"trainerId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
