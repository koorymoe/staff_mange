// موديل جدول "SupplierSpecialtyLink" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SupplierSpecialtyLink",
		Columns: []Column{
			{Name: "supplierId", Type: "text", NotNull: true},
			{Name: "specialtyId", Type: "text", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "SupplierSpecialtyLink_pkey", Def: "PRIMARY KEY (\"supplierId\", \"specialtyId\")"},
			{Name: "SupplierSpecialtyLink_specialtyId_fkey", Def: "FOREIGN KEY (\"specialtyId\") REFERENCES \"SupplierSpecialty\"(id) ON DELETE CASCADE"},
			{Name: "SupplierSpecialtyLink_supplierId_fkey", Def: "FOREIGN KEY (\"supplierId\") REFERENCES \"Supplier\"(id) ON DELETE CASCADE"},
		},
	})
}
