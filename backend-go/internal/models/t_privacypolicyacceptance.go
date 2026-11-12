// موديل جدول "PrivacyPolicyAcceptance" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PrivacyPolicyAcceptance",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "acceptedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "pointsVersion", Type: "integer", NotNull: true, Default: "0"},
			{Name: "pointsHash", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "PrivacyPolicyAcceptance_employeeId_key", Def: "UNIQUE (\"employeeId\")"},
			{Name: "PrivacyPolicyAcceptance_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
