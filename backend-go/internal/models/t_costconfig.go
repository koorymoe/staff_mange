// موديل جدول "CostConfig" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-24 دازاي: «تطلع رواتب كل الموظفين من الحجوزات وزيادة» — الزيادة (ربح الشركة) نسبة فوگ التكاليف.
func init() {
	register(Table{
		Name: "CostConfig",
		Columns: []Column{
			{Name: "id", Type: "integer", NotNull: true, Default: "1"},
			{Name: "marginPct", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "CostConfig_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "cost_config_single_row", Def: "CHECK ((id = 1))"},
			{Name: "cost_config_margin_ok", Def: "CHECK (((\"marginPct\" >= (0)::numeric) AND (\"marginPct\" <= (500)::numeric)))"},
		},
	})
}
