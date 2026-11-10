// موديل جدول "Exhibition" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Exhibition",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "location", Type: "text", NotNull: true},
			{Name: "startDate", Type: "text", NotNull: true},
			{Name: "endDate", Type: "text", NotNull: true},
			{Name: "companies", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "productsToShow", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "nominatedEmployeeIds", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "businessCardPhotos", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "keyFindings", Type: "text"},
			{Name: "visitReport", Type: "text"},
			{Name: "archived", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "createdById", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "Exhibition_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "Exhibition_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Exhibition_archived_idx\" ON public.\"Exhibition\" USING btree (archived)",
		},
	})
}
