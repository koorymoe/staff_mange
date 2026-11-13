// موديل جدول "MatrixOrder" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixOrder",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "actorId", Type: "text", NotNull: true},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PROPOSED'::text"},
			{Name: "actions", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "reply", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "doneAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "MatrixOrder_actorId_fkey", Def: "FOREIGN KEY (\"actorId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "MatrixOrder_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
