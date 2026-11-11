// موديل جدول "DesignSet" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DesignSet",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "projectId", Type: "text"},
			{Name: "customerId", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "archivedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "DesignSet_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "DesignSet_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE SET NULL"},
			{Name: "DesignSet_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DesignSet_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE SET NULL"},
		},
	})
}
