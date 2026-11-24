// موديل جدول "MoneySplitConfig" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-24 دازاي: «توزيع الفلوس» — نسب يحددها المالك من الإعدادات بدل البرمجة.
//   - stats*: الإحصائيات (شكد دخّل كل موظف): الليدر، والفنيين (يتقاسمون)، والحجز البلا ليدر.
//   - revenue*: الإيراد (صافي ربح الشركة): الليدر، والفني، والشركة الباقي.
func init() {
	register(Table{
		Name: "MoneySplitConfig",
		Columns: []Column{
			{Name: "id", Type: "integer", NotNull: true, Default: "1"},
			{Name: "statsLeaderPct", Type: "numeric", NotNull: true, Default: "50"},
			{Name: "statsTechPct", Type: "numeric", NotNull: true, Default: "50"},
			{Name: "statsNoLeaderTechPct", Type: "numeric", NotNull: true, Default: "30"},
			{Name: "revenueLeaderPct", Type: "numeric", NotNull: true, Default: "40"},
			{Name: "revenueTechPct", Type: "numeric", NotNull: true, Default: "30"},
			{Name: "revenueTechEach", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "MoneySplitConfig_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "money_split_single_row", Def: "CHECK ((id = 1))"},
			{Name: "MoneySplitConfig_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
