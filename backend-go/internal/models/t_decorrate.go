// موديل جدول "DecorRate" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DecorRate",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "unit", Type: "text", NotNull: true},
			{Name: "price", Type: "numeric", NotNull: true},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DecorRate_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DecorRate_price_check", Def: "CHECK ((price >= (0)::numeric))"},
			{Name: "DecorRate_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"DecorRate_service_idx\" ON public.\"DecorRate\" USING btree (\"serviceId\")",
		},
	})
}
