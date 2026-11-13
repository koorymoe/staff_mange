// موديل جدول "PermissionRevoke" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.
// (دازاي 11-13) صلاحية شالها الأدمن/المالك بيده من موظف: صلاحيات الدور الافتراضية
// (كل ٦ ساعات) ومنصب الهيكلة ما يرجّعوها. المنح اليدوي بعدين يمسح الصف.

package models

func init() {
	register(Table{
		Name: "PermissionRevoke",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "permission", Type: "text", NotNull: true},
			{Name: "revokedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "PermissionRevoke_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "PermissionRevoke_pkey", Def: "PRIMARY KEY (\"employeeId\", permission)"},
		},
	})
}
