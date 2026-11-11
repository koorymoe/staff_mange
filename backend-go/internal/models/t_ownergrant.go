// موديل جدول "OwnerGrant" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OwnerGrant",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "action", Type: "text", NotNull: true},
			{Name: "grantedById", Type: "text"},
			{Name: "grantedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "OwnerGrant_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "OwnerGrant_pkey", Def: "PRIMARY KEY (\"employeeId\", action)"},
		},
	})
}
