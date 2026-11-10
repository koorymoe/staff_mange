// موديل جدول "TechShowcaseItem" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TechShowcaseItem",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "description", Type: "text"},
			{Name: "mediaUrls", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "TechShowcaseItem_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "TechShowcaseItem_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"TechShowcaseItem_employeeId_idx\" ON public.\"TechShowcaseItem\" USING btree (\"employeeId\")",
		},
	})
}
