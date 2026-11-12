// موديل جدول "ProcurementRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProcurementRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "code", Type: "text", NotNull: true},
			{Name: "requestedById", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "status", Type: "\"ProcurementRequestStatus\"", NotNull: true, Default: "'PENDING'::\"ProcurementRequestStatus\""},
			{Name: "fulfilledById", Type: "text"},
			{Name: "totalCost", Type: "double precision"},
			{Name: "fulfillmentNotes", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "fulfilledAt", Type: "timestamp without time zone"},
			{Name: "requestType", Type: "text", NotNull: true, Default: "'CUSTOMER_PRODUCT'::text"},
			{Name: "supplierId", Type: "text"},
			{Name: "decidedAt", Type: "timestamp without time zone"},
			{Name: "decidedById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "ProcurementRequest_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id)"},
			{Name: "ProcurementRequest_code_key", Def: "UNIQUE (code)"},
			{Name: "ProcurementRequest_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ProcurementRequest_fulfilledById_fkey", Def: "FOREIGN KEY (\"fulfilledById\") REFERENCES \"Employee\"(id)"},
			{Name: "ProcurementRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProcurementRequest_requestedById_fkey", Def: "FOREIGN KEY (\"requestedById\") REFERENCES \"Employee\"(id)"},
			{Name: "ProcurementRequest_supplierId_fkey", Def: "FOREIGN KEY (\"supplierId\") REFERENCES \"Supplier\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ProcurementRequest_requestedById_idx\" ON public.\"ProcurementRequest\" USING btree (\"requestedById\")",
			"CREATE INDEX \"ProcurementRequest_status_idx\" ON public.\"ProcurementRequest\" USING btree (status)",
		},
	})
}
