// موديل جدول "AiBrainBudget" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiBrainBudget",
		Columns: []Column{
			{Name: "brain", Type: "text", NotNull: true},
			{Name: "dailyUSD", Type: "double precision", NotNull: true},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiBrainBudget_brain_check", Def: "CHECK ((brain = ANY (ARRAY['internal'::text, 'external'::text])))"},
			{Name: "AiBrainBudget_dailyUSD_check", Def: "CHECK ((\"dailyUSD\" > (0)::double precision))"},
			{Name: "AiBrainBudget_pkey", Def: "PRIMARY KEY (brain)"},
		},
	})
}
