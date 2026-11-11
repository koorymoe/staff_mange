// موديل جدول "QuotationTemplate" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "QuotationTemplate",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "sourceQuotationId", Type: "text"},
			{Name: "data", Type: "jsonb", NotNull: true},
			{Name: "styleNotes", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "archivedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "QuotationTemplate_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "QuotationTemplate_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "QuotationTemplate_sourceQuotationId_fkey", Def: "FOREIGN KEY (\"sourceQuotationId\") REFERENCES \"Quotation\"(id) ON DELETE SET NULL"},
		},
	})
}
