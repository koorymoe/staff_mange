// موديل جدول "KpiCriterion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "KpiCriterion",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "label", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "clause", Type: "integer"},
		},
		Constraints: []Constraint{
			{Name: "KpiCriterion_label_key", Def: "UNIQUE (label)"},
			{Name: "KpiCriterion_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
