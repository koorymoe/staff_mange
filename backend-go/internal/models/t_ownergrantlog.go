// موديل جدول "OwnerGrantLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OwnerGrantLog",
		Columns: []Column{
			{Name: "id", Type: "bigint", NotNull: true, Default: "nextval('\"OwnerGrantLog_id_seq\"'::regclass)"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "action", Type: "text", NotNull: true},
			{Name: "granted", Type: "boolean", NotNull: true},
			{Name: "byId", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "OwnerGrantLog_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
