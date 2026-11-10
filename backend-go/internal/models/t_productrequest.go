// موديل جدول "ProductRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProductRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "requestedById", Type: "text", NotNull: true},
			{Name: "productName", Type: "text", NotNull: true},
			{Name: "specs", Type: "text"},
			{Name: "source", Type: "text"},
			{Name: "model", Type: "text"},
			{Name: "category", Type: "text"},
			{Name: "price", Type: "double precision"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "resolvedAt", Type: "timestamp without time zone"},
			{Name: "resolvedById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "ProductRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProductRequest_requestedById_fkey", Def: "FOREIGN KEY (\"requestedById\") REFERENCES \"Employee\"(id)"},
			{Name: "ProductRequest_resolvedById_fkey", Def: "FOREIGN KEY (\"resolvedById\") REFERENCES \"Employee\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ProductRequest_status_idx\" ON public.\"ProductRequest\" USING btree (status)",
		},
	})
}
