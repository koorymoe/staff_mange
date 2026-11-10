// موديل جدول "DisciplineConfig" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DisciplineConfig",
		Columns: []Column{
			{Name: "id", Type: "integer", NotNull: true, Default: "1"},
			{Name: "startsAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DisciplineConfig_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "discipline_config_single_row", Def: "CHECK ((id = 1))"},
		},
	})
}
