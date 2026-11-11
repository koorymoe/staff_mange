// موديل جدول "DeletedProject" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DeletedProject",
		Columns: []Column{
			{Name: "id", Type: "bigint", NotNull: true, Default: "nextval('\"DeletedProject_id_seq\"'::regclass)"},
			{Name: "projectId", Type: "text", NotNull: true},
			{Name: "code", Type: "text"},
			{Name: "name", Type: "text"},
			{Name: "snapshot", Type: "jsonb", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "deletedById", Type: "text"},
			{Name: "deletedByName", Type: "text"},
			{Name: "deletedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DeletedProject_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
