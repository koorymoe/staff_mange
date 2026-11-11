// موديل جدول "QuotationTerm" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "QuotationTerm",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "sortOrder", Type: "integer", NotNull: true, Default: "0"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "(now() AT TIME ZONE 'UTC'::text)"},
		},
		Constraints: []Constraint{
			{Name: "QuotationTerm_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "QuotationTerm_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
