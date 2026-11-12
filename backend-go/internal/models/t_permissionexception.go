// موديل جدول "PermissionException" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PermissionException",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "permission", Type: "text", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "approvedById", Type: "text"},
			{Name: "expiresAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "PermissionException_approvedById_fkey", Def: "FOREIGN KEY (\"approvedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "PermissionException_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "PermissionException_pkey", Def: "PRIMARY KEY (\"employeeId\", permission)"},
		},
	})
}
