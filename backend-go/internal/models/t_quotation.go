// موديل جدول "Quotation" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Quotation",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "quotationNumber", Type: "text", NotNull: true},
			{Name: "customerName", Type: "text", NotNull: true},
			{Name: "customerPhone", Type: "text"},
			{Name: "customerAddress", Type: "text"},
			{Name: "projectName", Type: "text"},
			{Name: "grandTotal", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "discountPercent", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "discountValue", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "netTotal", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "notes", Type: "text"},
			{Name: "duration", Type: "text"},
			{Name: "status", Type: "\"QuotationStatus\"", NotNull: true, Default: "'NEW'::\"QuotationStatus\""},
			{Name: "createdByEmployeeId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "projectId", Type: "text"},
			{Name: "terms", Type: "jsonb"},
			{Name: "layout", Type: "text", NotNull: true, Default: "'CLASSIC'::text"},
			{Name: "cover", Type: "jsonb"},
			{Name: "annex", Type: "jsonb"},
			{Name: "validUntil", Type: "date"},
			{Name: "expiryRemindedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "Quotation_createdByEmployeeId_fkey", Def: "FOREIGN KEY (\"createdByEmployeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "Quotation_layout_check", Def: "CHECK ((layout = ANY (ARRAY['CLASSIC'::text, 'PROJECT'::text])))"},
			{Name: "Quotation_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "Quotation_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE SET NULL"},
			{Name: "Quotation_quotationNumber_key", Def: "UNIQUE (\"quotationNumber\")"},
		},
		Indexes: []string{
			"CREATE INDEX \"Quotation_createdByEmployeeId_idx\" ON public.\"Quotation\" USING btree (\"createdByEmployeeId\")",
			"CREATE INDEX \"Quotation_customerName_trgm_idx\" ON public.\"Quotation\" USING gin (ar_norm(\"customerName\") gin_trgm_ops)",
			"CREATE INDEX \"Quotation_projectId_idx\" ON public.\"Quotation\" USING btree (\"projectId\")",
			"CREATE INDEX \"Quotation_status_idx\" ON public.\"Quotation\" USING btree (status)",
		},
	})
}
