// موديل جدول "OrgDirectiveAck" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgDirectiveAck",
		Columns: []Column{
			{Name: "directiveId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "ackAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "OrgDirectiveAck_directiveId_fkey", Def: "FOREIGN KEY (\"directiveId\") REFERENCES \"OrgDirective\"(id) ON DELETE CASCADE"},
			{Name: "OrgDirectiveAck_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "OrgDirectiveAck_pkey", Def: "PRIMARY KEY (\"directiveId\", \"employeeId\")"},
		},
	})
}
