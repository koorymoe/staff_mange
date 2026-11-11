// موديل جدول "SystemPriceCatalog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SystemPriceCatalog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "systemName", Type: "text", NotNull: true},
			{Name: "itemName", Type: "text", NotNull: true},
			{Name: "category", Type: "text", NotNull: true},
			{Name: "value", Type: "numeric", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "SystemPriceCatalog_category_check", Def: "CHECK ((category = ANY (ARRAY['install'::text, 'wiring'::text, 'programming'::text])))"},
			{Name: "SystemPriceCatalog_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SystemPriceCatalog_systemName_itemName_category_key", Def: "UNIQUE (\"systemName\", \"itemName\", category)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SystemPriceCatalog_systemName_idx\" ON public.\"SystemPriceCatalog\" USING btree (\"systemName\")",
		},
	})
}
