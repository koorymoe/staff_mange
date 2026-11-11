// موديل جدول "MatrixLesson" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixLesson",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "context", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "agent", Type: "text", NotNull: true, Default: "'MATRIX'::text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'ACTIVE'::text"},
		},
		Constraints: []Constraint{
			{Name: "MatrixLesson_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "MatrixLesson_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MatrixLesson_status_check", Def: "CHECK ((status = ANY (ARRAY['ACTIVE'::text, 'PROPOSED'::text])))"},
		},
	})
}
