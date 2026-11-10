// موديل جدول "NetworkPriceItem" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "NetworkPriceItem",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "label", Type: "text", NotNull: true},
			{Name: "unit", Type: "text", NotNull: true, Default: "'قطعة'::text"},
			{Name: "pricingMode", Type: "text", NotNull: true, Default: "'FLAT'::text"},
			{Name: "basePrice", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "includedQty", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "extraPerUnit", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "brackets", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "note", Type: "text"},
			{Name: "sortOrder", Type: "integer", NotNull: true, Default: "0"},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "NetworkPriceItem_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"NetworkPriceItem_active_idx\" ON public.\"NetworkPriceItem\" USING btree (active, \"sortOrder\")",
		},
	})
}
