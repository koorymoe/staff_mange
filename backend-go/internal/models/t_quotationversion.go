// موديل جدول "QuotationVersion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "QuotationVersion",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "quotationId", Type: "text", NotNull: true},
			{Name: "version", Type: "integer", NotNull: true},
			{Name: "snapshot", Type: "jsonb", NotNull: true},
			{Name: "archivedById", Type: "text"},
			{Name: "archivedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "QuotationVersion_archivedById_fkey", Def: "FOREIGN KEY (\"archivedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "QuotationVersion_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "QuotationVersion_quotationId_fkey", Def: "FOREIGN KEY (\"quotationId\") REFERENCES \"Quotation\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"QuotationVersion_uniq\" ON public.\"QuotationVersion\" USING btree (\"quotationId\", version)",
		},
	})
}
