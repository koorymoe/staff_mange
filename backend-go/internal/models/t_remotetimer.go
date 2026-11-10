// موديل جدول "RemoteTimer" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "RemoteTimer",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "startedAt", Type: "timestamp with time zone"},
			{Name: "accumulated", Type: "integer", NotNull: true, Default: "0"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "RemoteTimer_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "RemoteTimer_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
		},
	})
}
