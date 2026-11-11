// موديل جدول "CreditPayment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CreditPayment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "creditId", Type: "text", NotNull: true},
			{Name: "amount", Type: "numeric", NotNull: true},
			{Name: "paidAt", Type: "date", NotNull: true, Default: "CURRENT_DATE"},
			{Name: "note", Type: "text"},
			{Name: "byId", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "CreditPayment_amount_check", Def: "CHECK ((amount > (0)::numeric))"},
			{Name: "CreditPayment_creditId_fkey", Def: "FOREIGN KEY (\"creditId\") REFERENCES \"CreditSale\"(id) ON DELETE CASCADE"},
			{Name: "CreditPayment_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX credit_payment_credit ON public.\"CreditPayment\" USING btree (\"creditId\")",
		},
	})
}
