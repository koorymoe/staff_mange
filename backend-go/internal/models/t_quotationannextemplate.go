// موديل جدول "QuotationAnnexTemplate" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "QuotationAnnexTemplate",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "parts", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "sortOrder", Type: "integer", NotNull: true, Default: "0"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "QuotationAnnexTemplate_name_key", Def: "UNIQUE (name)"},
			{Name: "QuotationAnnexTemplate_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "QuotationAnnexTemplate_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
