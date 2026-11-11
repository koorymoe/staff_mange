// موديل جدول "PushVapid" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PushVapid",
		Columns: []Column{
			{Name: "id", Type: "integer", NotNull: true},
			{Name: "publicKey", Type: "text", NotNull: true},
			{Name: "vapidSecret", Type: "text", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "PushVapid_id_check", Def: "CHECK ((id = 1))"},
			{Name: "PushVapid_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
