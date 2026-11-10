// موديل جدول "ProjectContractValue" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectContractValue",
		Columns: []Column{
			{Name: "projectId", Type: "text", NotNull: true},
			{Name: "amount", Type: "numeric(16,0)", NotNull: true},
			{Name: "setById", Type: "text"},
			{Name: "setAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ProjectContractValue_amount_check", Def: "CHECK ((amount >= (0)::numeric))"},
			{Name: "ProjectContractValue_pkey", Def: "PRIMARY KEY (\"projectId\")"},
			{Name: "ProjectContractValue_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE CASCADE"},
			{Name: "ProjectContractValue_setById_fkey", Def: "FOREIGN KEY (\"setById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
