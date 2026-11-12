// موديل جدول "DesignAsset" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DesignAsset",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "category", Type: "text", NotNull: true},
			{Name: "notes", Type: "text"},
			{Name: "fileKey", Type: "text", NotNull: true},
			{Name: "fileType", Type: "text"},
			{Name: "uploadedById", Type: "text"},
			{Name: "uploadedByName", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "archivedAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "DesignAsset_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DesignAsset_uploadedById_fkey", Def: "FOREIGN KEY (\"uploadedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"DesignAsset_live_idx\" ON public.\"DesignAsset\" USING btree (\"createdAt\" DESC) WHERE (\"archivedAt\" IS NULL)",
		},
	})
}
