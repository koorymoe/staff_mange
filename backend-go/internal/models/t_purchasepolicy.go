// موديل جدول "PurchasePolicy" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PurchasePolicy",
		Columns: []Column{
			{Name: "id", Type: "integer", NotNull: true, Default: "1"},
			{Name: "quoteThreshold", Type: "double precision", NotNull: true, Default: "500000"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "ceilingPct", Type: "double precision", NotNull: true, Default: "10"},
			{Name: "contactDays", Type: "integer", NotNull: true, Default: "30"},
		},
		Constraints: []Constraint{
			{Name: "PurchasePolicy_id_check", Def: "CHECK ((id = 1))"},
			{Name: "PurchasePolicy_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
