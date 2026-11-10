// موديل جدول "SimContentVersion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimContentVersion",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "entityKind", Type: "text", NotNull: true},
			{Name: "entityId", Type: "text", NotNull: true},
			{Name: "version", Type: "integer", NotNull: true},
			{Name: "snapshot", Type: "jsonb", NotNull: true},
			{Name: "publishedById", Type: "text"},
			{Name: "publishedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimContentVersion_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SimContentVersion_publishedById_fkey", Def: "FOREIGN KEY (\"publishedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"SimContentVersion_uniq\" ON public.\"SimContentVersion\" USING btree (\"entityKind\", \"entityId\", version)",
		},
	})
}
