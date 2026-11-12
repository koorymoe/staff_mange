// موديل جدول "PlanFile" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PlanFile",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "setId", Type: "text", NotNull: true},
			{Name: "typeCode", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "fileKey", Type: "text", NotNull: true},
			{Name: "fileType", Type: "text"},
			{Name: "uploadedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "removedAt", Type: "timestamp with time zone"},
			{Name: "removedById", Type: "text"},
			{Name: "sheetPos", Type: "text", NotNull: true, Default: "'right'::text"},
			{Name: "sheetName", Type: "text"},
			{Name: "sheetNotes", Type: "text"},
			{Name: "sheetScale", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "PlanFile_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "PlanFile_removedById_fkey", Def: "FOREIGN KEY (\"removedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "PlanFile_setId_fkey", Def: "FOREIGN KEY (\"setId\") REFERENCES \"PlanSet\"(id) ON DELETE CASCADE"},
			{Name: "PlanFile_sheetPos_check", Def: "CHECK ((\"sheetPos\" = ANY (ARRAY['right'::text, 'bottom'::text])))"},
			{Name: "PlanFile_uploadedById_fkey", Def: "FOREIGN KEY (\"uploadedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"PlanFile_set_idx\" ON public.\"PlanFile\" USING btree (\"setId\") WHERE (\"removedAt\" IS NULL)",
		},
	})
}
