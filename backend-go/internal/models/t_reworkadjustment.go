// موديل جدول "ReworkAdjustment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-24: المبالغ الي انتقلت بموافقة المراقب (+ للي صلّح، − للكادر الأول) — تنحسب بـ«الفلوس الي دخّلها».
func init() {
	register(Table{
		Name: "ReworkAdjustment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "claimId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "amount", Type: "numeric", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ReworkAdjustment_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ReworkAdjustment_claimId_fkey", Def: "FOREIGN KEY (\"claimId\") REFERENCES \"ReworkClaim\"(id) ON DELETE CASCADE"},
			{Name: "ReworkAdjustment_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{"CREATE INDEX \"ReworkAdjustment_emp_idx\" ON public.\"ReworkAdjustment\" USING btree (\"employeeId\", \"createdAt\")"},
	})
}
