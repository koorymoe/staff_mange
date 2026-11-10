// موديل جدول "ProcurementVerify" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProcurementVerify",
		Columns: []Column{
			{Name: "procurementId", Type: "text", NotNull: true},
			{Name: "verifiedById", Type: "text", NotNull: true},
			{Name: "verifiedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "note", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "ProcurementVerify_pkey", Def: "PRIMARY KEY (\"procurementId\")"},
			{Name: "ProcurementVerify_procurementId_fkey", Def: "FOREIGN KEY (\"procurementId\") REFERENCES \"ProcurementRequest\"(id) ON DELETE CASCADE"},
		},
	})
}
