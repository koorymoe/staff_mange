// موديل جدول "LeaderInvoiceDeleteRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "LeaderInvoiceDeleteRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "invoiceId", Type: "text", NotNull: true},
			{Name: "accountingCode", Type: "text"},
			{Name: "bookingCode", Type: "text"},
			{Name: "netTotal", Type: "numeric"},
			{Name: "requestedById", Type: "text"},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp without time zone"},
			{Name: "decisionNote", Type: "text"},
			{Name: "snapshot", Type: "jsonb"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "LeaderInvoiceDeleteRequest_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "LeaderInvoiceDeleteRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "LeaderInvoiceDeleteRequest_requestedById_fkey", Def: "FOREIGN KEY (\"requestedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "LeaderInvoiceDeleteRequest_status_check", Def: "CHECK ((status = ANY (ARRAY['PENDING'::text, 'APPROVED'::text, 'REJECTED'::text])))"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"LeaderInvoiceDeleteRequest_pending\" ON public.\"LeaderInvoiceDeleteRequest\" USING btree (\"invoiceId\") WHERE (status = 'PENDING'::text)",
		},
	})
}
