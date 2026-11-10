// موديل جدول "RevolvingFund" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "RevolvingFund",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "balance", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "isActive", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "RevolvingFund_name_key", Def: "UNIQUE (name)"},
			{Name: "RevolvingFund_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
