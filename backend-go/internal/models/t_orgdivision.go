// موديل جدول "OrgDivision" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgDivision",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "parentId", Type: "text"},
			{Name: "sort", Type: "integer", NotNull: true, Default: "0"},
			{Name: "summary", Type: "text", NotNull: true, Default: "''::text"},
		},
		Constraints: []Constraint{
			{Name: "OrgDivision_parentId_fkey", Def: "FOREIGN KEY (\"parentId\") REFERENCES \"OrgDivision\"(id)"},
			{Name: "OrgDivision_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
