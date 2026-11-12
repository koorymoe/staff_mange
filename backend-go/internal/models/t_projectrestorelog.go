// موديل جدول "ProjectRestoreLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectRestoreLog",
		Columns: []Column{
			{Name: "id", Type: "bigint", NotNull: true, Default: "nextval('\"ProjectRestoreLog_id_seq\"'::regclass)"},
			{Name: "projectId", Type: "text", NotNull: true},
			{Name: "toStage", Type: "text", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "byId", Type: "text"},
			{Name: "byName", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ProjectRestoreLog_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProjectRestoreLog_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE CASCADE"},
		},
	})
}
