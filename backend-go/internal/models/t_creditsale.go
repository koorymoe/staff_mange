// موديل جدول "CreditSale" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CreditSale",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true, Default: "'BOOKING'::text"},
			{Name: "bookingId", Type: "text"},
			{Name: "projectId", Type: "text"},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "customerName", Type: "text"},
			{Name: "phone", Type: "text"},
			{Name: "amount", Type: "numeric", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "dueDate", Type: "date"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "settledAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "CreditSale_amount_check", Def: "CHECK ((amount > (0)::numeric))"},
			{Name: "CreditSale_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "CreditSale_bookingId_key", Def: "UNIQUE (\"bookingId\")"},
			{Name: "CreditSale_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "CreditSale_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE SET NULL"},
		},
	})
}
