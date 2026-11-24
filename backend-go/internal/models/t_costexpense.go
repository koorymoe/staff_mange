// موديل جدول "CostExpense" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-24 دازاي: مصاريف الشهر — الكاز، تصليح السيارات، العدد الجديدة، غيرها. إدخال يدوي.
func init() {
	register(Table{
		Name: "CostExpense",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "month", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "amount", Type: "numeric", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "CostExpense_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "CostExpense_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "cost_expense_kind_ok", Def: "CHECK ((kind = ANY (ARRAY['FUEL'::text, 'VEHICLE'::text, 'TOOLS'::text, 'OTHER'::text])))"},
			{Name: "cost_expense_amount_ok", Def: "CHECK ((amount >= (0)::numeric))"},
			{Name: "cost_expense_month_ok", Def: "CHECK ((month ~ '^[0-9]{4}-[0-9]{2}$'::text))"},
		},
		Indexes: []string{"CREATE INDEX \"CostExpense_month_idx\" ON public.\"CostExpense\" USING btree (month)"},
	})
}
