// موديل جدول "PlanSet" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PlanSet",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "projectId", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "archivedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "PlanSet_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "PlanSet_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "PlanSet_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE SET NULL"},
		},
	})
}
