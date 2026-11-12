// موديل جدول "WorkTypeImage" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "WorkTypeImage",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "workTypeId", Type: "text", NotNull: true},
			{Name: "fileKey", Type: "text", NotNull: true},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "useCount", Type: "integer", NotNull: true, Default: "0"},
			{Name: "lastUsedAt", Type: "timestamp with time zone"},
			{Name: "removedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "WorkTypeImage_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "WorkTypeImage_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "WorkTypeImage_workTypeId_fkey", Def: "FOREIGN KEY (\"workTypeId\") REFERENCES \"ProjectWorkType\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"WorkTypeImage_pick_idx\" ON public.\"WorkTypeImage\" USING btree (\"workTypeId\", \"useCount\") WHERE (\"removedAt\" IS NULL)",
		},
	})
}
