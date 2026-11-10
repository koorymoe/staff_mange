// موديل جدول "RevolvingFundTxn" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "RevolvingFundTxn",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "fundId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "amount", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "spentAmount", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "returnedAmount", Type: "numeric(14,2)", NotNull: true, Default: "0"},
			{Name: "bookingId", Type: "text"},
			{Name: "receiptImage", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'APPROVED'::text"},
			{Name: "createdById", Type: "text"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
			{Name: "reviewNote", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "dischargedAt", Type: "timestamp with time zone"},
			{Name: "dischargedById", Type: "text"},
			{Name: "dischargeAccount", Type: "text"},
			{Name: "dischargeNote", Type: "text"},
			{Name: "requestId", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "RevolvingFundTxn_dischargedById_fkey", Def: "FOREIGN KEY (\"dischargedById\") REFERENCES \"Employee\"(id)"},
			{Name: "RevolvingFundTxn_fundId_fkey", Def: "FOREIGN KEY (\"fundId\") REFERENCES \"RevolvingFund\"(id)"},
			{Name: "RevolvingFundTxn_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "fund_txn_nonneg", Def: "CHECK (((COALESCE(\"spentAmount\", (0)::numeric) >= (0)::numeric) AND (COALESCE(\"returnedAmount\", (0)::numeric) >= (0)::numeric))) NOT VALID"},
		},
		Indexes: []string{
			"CREATE INDEX \"RevolvingFundTxn_discharge_idx\" ON public.\"RevolvingFundTxn\" USING btree (kind, status, \"dischargedAt\")",
			"CREATE INDEX \"RevolvingFundTxn_employee_idx\" ON public.\"RevolvingFundTxn\" USING btree (\"employeeId\", \"createdAt\" DESC)",
			"CREATE INDEX \"RevolvingFundTxn_status_idx\" ON public.\"RevolvingFundTxn\" USING btree (status, \"createdAt\" DESC)",
		},
	})
}
