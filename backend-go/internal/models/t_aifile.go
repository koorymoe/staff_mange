// موديل جدول "AiFile" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiFile",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "data", Type: "bytea", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiFile_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "AiFile_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
