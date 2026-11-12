// موديل جدول "SolarComponent" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SolarComponent",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "category", Type: "text", NotNull: true},
			{Name: "quantity", Type: "integer", NotNull: true, Default: "0"},
			{Name: "price", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "minStock", Type: "integer", NotNull: true, Default: "0"},
			{Name: "specs", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "notes", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SolarComponent_category_check", Def: "CHECK ((category = ANY (ARRAY['PANEL'::text, 'INVERTER'::text, 'BATTERY'::text, 'BOARD'::text, 'IRON'::text])))"},
			{Name: "SolarComponent_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SolarComponent_category_idx\" ON public.\"SolarComponent\" USING btree (category)",
			"CREATE INDEX \"SolarComponent_low_stock_idx\" ON public.\"SolarComponent\" USING btree (quantity) WHERE (quantity <= 0)",
			"CREATE UNIQUE INDEX \"SolarComponent_name_category_key\" ON public.\"SolarComponent\" USING btree (lower(btrim(name)), category)",
		},
	})
}
