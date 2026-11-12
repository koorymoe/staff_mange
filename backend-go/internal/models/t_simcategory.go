// موديل جدول "SimCategory" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimCategory",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text"},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "description", Type: "text"},
			{Name: "imagePath", Type: "text"},
			{Name: "sortOrder", Type: "integer", NotNull: true, Default: "0"},
			{Name: "archived", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimCategory_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SimCategory_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"SimCategory_service_idx\" ON public.\"SimCategory\" USING btree (\"serviceId\", archived, \"sortOrder\")",
		},
	})
}
