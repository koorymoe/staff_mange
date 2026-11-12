// موديل جدول "SimLesson" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimLesson",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "categoryId", Type: "text"},
			{Name: "deviceId", Type: "text"},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "blocks", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "sortOrder", Type: "integer", NotNull: true, Default: "0"},
			{Name: "status", Type: "text", NotNull: true, Default: "'DRAFT'::text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimLesson_categoryId_fkey", Def: "FOREIGN KEY (\"categoryId\") REFERENCES \"SimCategory\"(id) ON DELETE CASCADE"},
			{Name: "SimLesson_deviceId_fkey", Def: "FOREIGN KEY (\"deviceId\") REFERENCES \"SimDevice\"(id) ON DELETE CASCADE"},
			{Name: "SimLesson_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SimLesson_cat_idx\" ON public.\"SimLesson\" USING btree (\"categoryId\", \"sortOrder\")",
		},
	})
}
