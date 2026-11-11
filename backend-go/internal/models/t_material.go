// موديل جدول "Material" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Material",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "code", Type: "text", NotNull: true},
			{Name: "sellPrice", Type: "numeric", NotNull: true},
			{Name: "profitPerUnit", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "wholesalePrice", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "stockQty", Type: "double precision"},
			{Name: "stockCountedAt", Type: "timestamp with time zone"},
			{Name: "stockCountedById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "Material_code_key", Def: "UNIQUE (code)"},
			{Name: "Material_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "Material_stockCountedById_fkey", Def: "FOREIGN KEY (\"stockCountedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"Material_code_idx\" ON public.\"Material\" USING btree (code)",
		},
	})
}
