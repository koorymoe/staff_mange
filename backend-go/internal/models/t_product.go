// موديل جدول "Product" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Product",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "unit", Type: "text"},
			{Name: "defaultPrice", Type: "double precision"},
			{Name: "imageBase64", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "wholesalePrice", Type: "numeric"},
			{Name: "availability", Type: "text", NotNull: true, Default: "'IN_STOCK'::text"},
			{Name: "serviceId", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "serviceText", Type: "text"},
			{Name: "specs", Type: "text"},
			{Name: "source", Type: "text"},
			{Name: "modelName", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "Product_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "Product_name_key", Def: "UNIQUE (name)"},
			{Name: "Product_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "Product_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"Product_created_by_idx\" ON public.\"Product\" USING btree (\"createdById\")",
			"CREATE INDEX \"Product_service_idx\" ON public.\"Product\" USING btree (\"serviceId\")",
		},
	})
}
