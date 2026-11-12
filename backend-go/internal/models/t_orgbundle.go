// موديل جدول "OrgBundle" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgBundle",
		Columns: []Column{
			{Name: "scope", Type: "text", NotNull: true},
			{Name: "permission", Type: "text", NotNull: true},
			{Name: "headOnly", Type: "boolean", NotNull: true, Default: "false"},
		},
		Constraints: []Constraint{
			{Name: "OrgBundle_pkey", Def: "PRIMARY KEY (scope, permission)"},
		},
	})
}
