// موديل جدول "ProductProcurement" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProductProcurement",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "requestId", Type: "text", NotNull: true},
			{Name: "fundId", Type: "text", NotNull: true},
			{Name: "supplierId", Type: "text", NotNull: true},
			{Name: "productName", Type: "text", NotNull: true},
			{Name: "spentAmount", Type: "numeric(14,2)", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "receiptImage", Type: "text"},
			{Name: "payerKind", Type: "text", NotNull: true},
			{Name: "customerNote", Type: "text"},
			{Name: "bookingId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "purchasedById", Type: "text", NotNull: true},
			{Name: "settledById", Type: "text"},
			{Name: "settledAt", Type: "timestamp with time zone"},
			{Name: "settleNote", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ProductProcurement_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id)"},
			{Name: "ProductProcurement_fundId_fkey", Def: "FOREIGN KEY (\"fundId\") REFERENCES \"RevolvingFund\"(id)"},
			{Name: "ProductProcurement_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProductProcurement_purchasedById_fkey", Def: "FOREIGN KEY (\"purchasedById\") REFERENCES \"Employee\"(id)"},
			{Name: "ProductProcurement_requestId_fkey", Def: "FOREIGN KEY (\"requestId\") REFERENCES \"ProductRequest\"(id) ON DELETE CASCADE"},
			{Name: "ProductProcurement_settledById_fkey", Def: "FOREIGN KEY (\"settledById\") REFERENCES \"Employee\"(id)"},
			{Name: "ProductProcurement_supplierId_fkey", Def: "FOREIGN KEY (\"supplierId\") REFERENCES \"Supplier\"(id)"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"ProductProcurement_one_per_request\" ON public.\"ProductProcurement\" USING btree (\"requestId\")",
			"CREATE INDEX \"ProductProcurement_status_idx\" ON public.\"ProductProcurement\" USING btree (status, \"createdAt\" DESC)",
		},
	})
}
