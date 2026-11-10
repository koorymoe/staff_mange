// موديل جدول "TechSupplier" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TechSupplier",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "companyName", Type: "text", NotNull: true},
			{Name: "ownerName", Type: "text"},
			{Name: "phone", Type: "text", NotNull: true},
			{Name: "address", Type: "text"},
			{Name: "locationUrl", Type: "text"},
			{Name: "specialty", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "TechSupplier_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "TechSupplier_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
