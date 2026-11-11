// موديل جدول "ServicePriceSuggestion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ServicePriceSuggestion",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "avgAmount", Type: "double precision", NotNull: true},
			{Name: "sampleCount", Type: "integer", NotNull: true},
			{Name: "minAmount", Type: "double precision", NotNull: true},
			{Name: "maxAmount", Type: "double precision", NotNull: true},
			{Name: "source", Type: "text", NotNull: true, Default: "'MANUAL_AVERAGE'::text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PROPOSED'::text"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
			{Name: "reviewNote", Type: "text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ServicePriceSuggestion_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ServicePriceSuggestion_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ServicePriceSuggestion_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ServicePriceSuggestion_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"ServicePriceSuggestion_open_uniq\" ON public.\"ServicePriceSuggestion\" USING btree (\"serviceId\") WHERE (status = ANY (ARRAY['PROPOSED'::text, 'REVIEWED'::text]))",
		},
	})
}
