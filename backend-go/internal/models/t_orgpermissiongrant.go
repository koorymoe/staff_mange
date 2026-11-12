// موديل جدول "OrgPermissionGrant" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgPermissionGrant",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "permission", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "OrgPermissionGrant_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "OrgPermissionGrant_pkey", Def: "PRIMARY KEY (\"employeeId\", permission)"},
		},
	})
}
