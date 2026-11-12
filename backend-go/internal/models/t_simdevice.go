// موديل جدول "SimDevice" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimDevice",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "categoryId", Type: "text", NotNull: true},
			{Name: "brand", Type: "text", NotNull: true},
			{Name: "model", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "summary", Type: "text"},
			{Name: "imagePath", Type: "text"},
			{Name: "engineKind", Type: "text", NotNull: true, Default: "'WIRING'::text"},
			{Name: "spec", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "terminals", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "ui", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "status", Type: "text", NotNull: true, Default: "'DRAFT'::text"},
			{Name: "version", Type: "integer", NotNull: true, Default: "1"},
			{Name: "sourceRef", Type: "text"},
			{Name: "localPractice", Type: "text"},
			{Name: "verified", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "verifiedById", Type: "text"},
			{Name: "verifiedAt", Type: "timestamp with time zone"},
			{Name: "authorId", Type: "text"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
			{Name: "reviewNote", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "geometry", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
		},
		Constraints: []Constraint{
			{Name: "SimDevice_authorId_fkey", Def: "FOREIGN KEY (\"authorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "SimDevice_categoryId_fkey", Def: "FOREIGN KEY (\"categoryId\") REFERENCES \"SimCategory\"(id) ON DELETE CASCADE"},
			{Name: "SimDevice_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SimDevice_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "SimDevice_verifiedById_fkey", Def: "FOREIGN KEY (\"verifiedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"SimDevice_category_idx\" ON public.\"SimDevice\" USING btree (\"categoryId\", status, verified)",
		},
	})
}
