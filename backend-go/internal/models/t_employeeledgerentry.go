// موديل جدول "EmployeeLedgerEntry" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeLedgerEntry",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "amount", Type: "numeric(14,2)", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "invoiceId", Type: "text"},
			{Name: "method", Type: "text"},
			{Name: "reference", Type: "text"},
			{Name: "occurredAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "recordedById", Type: "text"},
			{Name: "note", Type: "text"},
			{Name: "reversesEntryId", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeLedgerEntry_amount_check", Def: "CHECK ((amount > (0)::numeric))"},
			{Name: "EmployeeLedgerEntry_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id)"},
			{Name: "EmployeeLedgerEntry_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "EmployeeLedgerEntry_invoiceId_fkey", Def: "FOREIGN KEY (\"invoiceId\") REFERENCES \"LeaderInvoice\"(id) ON DELETE SET NULL"},
			{Name: "EmployeeLedgerEntry_kind_check", Def: "CHECK ((kind = ANY (ARRAY['COLLECTED'::text, 'HANDED_OVER'::text, 'ADJUSTMENT'::text, 'WRITE_OFF'::text])))"},
			{Name: "EmployeeLedgerEntry_method_check", Def: "CHECK ((method = ANY (ARRAY['CASH'::text, 'CARD'::text, 'TRANSFER'::text])))"},
			{Name: "EmployeeLedgerEntry_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "EmployeeLedgerEntry_recordedById_fkey", Def: "FOREIGN KEY (\"recordedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "EmployeeLedgerEntry_reversesEntryId_fkey", Def: "FOREIGN KEY (\"reversesEntryId\") REFERENCES \"EmployeeLedgerEntry\"(id)"},
			{Name: "employee_ledger_collected_needs_booking", Def: "CHECK (((kind <> 'COLLECTED'::text) OR (\"bookingId\" IS NOT NULL)))"},
			{Name: "employee_ledger_reversal_needs_note", Def: "CHECK (((\"reversesEntryId\" IS NULL) OR ((note IS NOT NULL) AND (btrim(note) <> ''::text))))"},
		},
		Indexes: []string{
			"CREATE INDEX \"EmployeeLedgerEntry_booking_idx\" ON public.\"EmployeeLedgerEntry\" USING btree (\"bookingId\")",
			"CREATE INDEX \"EmployeeLedgerEntry_employee_idx\" ON public.\"EmployeeLedgerEntry\" USING btree (\"employeeId\", \"occurredAt\" DESC)",
			"CREATE UNIQUE INDEX \"EmployeeLedgerEntry_reverse_unique\" ON public.\"EmployeeLedgerEntry\" USING btree (\"reversesEntryId\") WHERE (\"reversesEntryId\" IS NOT NULL)",
		},
	})
}
