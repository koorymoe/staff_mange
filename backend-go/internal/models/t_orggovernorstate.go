// موديل جدول "OrgGovernorState" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgGovernorState",
		Columns: []Column{
			{Name: "key", Type: "text", NotNull: true},
			{Name: "value", Type: "text", NotNull: true},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "OrgGovernorState_pkey", Def: "PRIMARY KEY (key)"},
		},
	})
}
