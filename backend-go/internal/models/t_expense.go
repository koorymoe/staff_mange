// موديل جدول "Expense" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Expense",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "amount", Type: "double precision", NotNull: true},
			{Name: "description", Type: "text"},
			{Name: "status", Type: "\"ExpenseStatus\"", NotNull: true, Default: "'PENDING'::\"ExpenseStatus\""},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "bookingId", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "Expense_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "Expense_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "Expense_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Expense_bookingId_idx\" ON public.\"Expense\" USING btree (\"bookingId\") WHERE (\"bookingId\" IS NOT NULL)",
			"CREATE INDEX \"Expense_employeeId_idx\" ON public.\"Expense\" USING btree (\"employeeId\")",
		},
	})
}
