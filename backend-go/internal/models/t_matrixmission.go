// موديل جدول "MatrixMission" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixMission",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "createdById", Type: "text", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "startsOn", Type: "date", NotNull: true},
			{Name: "endsOn", Type: "date", NotNull: true},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "summary", Type: "text"},
			{Name: "summaryAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "MatrixMission_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "MatrixMission_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "MatrixMission_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
