// موديل جدول "DesignFile" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DesignFile",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "sectionId", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "fileKey", Type: "text"},
			{Name: "fileType", Type: "text"},
			{Name: "url", Type: "text"},
			{Name: "toolName", Type: "text"},
			{Name: "uploadedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "removedAt", Type: "timestamp with time zone"},
			{Name: "removedById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "DesignFile_check", Def: "CHECK (((\"fileKey\" IS NULL) <> (url IS NULL)))"},
			{Name: "DesignFile_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DesignFile_removedById_fkey", Def: "FOREIGN KEY (\"removedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "DesignFile_sectionId_fkey", Def: "FOREIGN KEY (\"sectionId\") REFERENCES \"DesignSection\"(id) ON DELETE CASCADE"},
			{Name: "DesignFile_uploadedById_fkey", Def: "FOREIGN KEY (\"uploadedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"DesignFile_section_idx\" ON public.\"DesignFile\" USING btree (\"sectionId\")",
		},
	})
}
