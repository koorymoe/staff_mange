// موديل جدول "TechSupplierAccess" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TechSupplierAccess",
		Columns: []Column{
			{Name: "supplierId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "TechSupplierAccess_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "TechSupplierAccess_pkey", Def: "PRIMARY KEY (\"supplierId\", \"employeeId\")"},
			{Name: "TechSupplierAccess_supplierId_fkey", Def: "FOREIGN KEY (\"supplierId\") REFERENCES \"TechSupplier\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"TechSupplierAccess_emp_idx\" ON public.\"TechSupplierAccess\" USING btree (\"employeeId\")",
		},
	})
}
