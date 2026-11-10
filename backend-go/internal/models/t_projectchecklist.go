// موديل جدول "ProjectChecklist" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectChecklist",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "projectId", Type: "text"},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "createdById", Type: "text", NotNull: true},
			{Name: "photoUrls", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ProjectChecklist_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "ProjectChecklist_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProjectChecklist_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ProjectChecklist_projectId_idx\" ON public.\"ProjectChecklist\" USING btree (\"projectId\")",
		},
	})
}
