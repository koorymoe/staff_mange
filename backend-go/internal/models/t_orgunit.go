// موديل جدول "OrgUnit" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgUnit",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "divisionId", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "sort", Type: "integer", NotNull: true, Default: "0"},
		},
		Constraints: []Constraint{
			{Name: "OrgUnit_divisionId_fkey", Def: "FOREIGN KEY (\"divisionId\") REFERENCES \"OrgDivision\"(id)"},
			{Name: "OrgUnit_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
