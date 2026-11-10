// موديل جدول "ExpensePurchase" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ExpensePurchase",
		Columns: []Column{
			{Name: "expenseId", Type: "text", NotNull: true},
			{Name: "techSupplierId", Type: "text"},
			{Name: "productName", Type: "text", NotNull: true},
			{Name: "quantity", Type: "integer", NotNull: true},
			{Name: "unitPrice", Type: "double precision", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ExpensePurchase_expenseId_fkey", Def: "FOREIGN KEY (\"expenseId\") REFERENCES \"Expense\"(id) ON DELETE CASCADE"},
			{Name: "ExpensePurchase_pkey", Def: "PRIMARY KEY (\"expenseId\")"},
			{Name: "ExpensePurchase_techSupplierId_fkey", Def: "FOREIGN KEY (\"techSupplierId\") REFERENCES \"TechSupplier\"(id) ON DELETE SET NULL"},
		},
	})
}
