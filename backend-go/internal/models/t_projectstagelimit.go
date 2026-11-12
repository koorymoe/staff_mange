// موديل جدول "ProjectStageLimit" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectStageLimit",
		Columns: []Column{
			{Name: "stage", Type: "text", NotNull: true},
			{Name: "days", Type: "integer", NotNull: true},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ProjectStageLimit_days_check", Def: "CHECK ((days > 0))"},
			{Name: "ProjectStageLimit_pkey", Def: "PRIMARY KEY (stage)"},
			{Name: "ProjectStageLimit_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
