// موديل جدول "VipCustomer" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VipCustomer",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "requestSummary", Type: "text"},
			{Name: "note", Type: "text"},
			{Name: "markedByEmployeeId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "customerPosition", Type: "text"},
			{Name: "projectId", Type: "text"},
			{Name: "source", Type: "text", NotNull: true, Default: "'MANUAL'::text"},
			{Name: "boughtFromUs", Type: "boolean", NotNull: true, Default: "true"},
		},
		Constraints: []Constraint{
			{Name: "VipCustomer_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "VipCustomer_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE CASCADE"},
			{Name: "VipCustomer_customerId_key", Def: "UNIQUE (\"customerId\")"},
			{Name: "VipCustomer_markedByEmployeeId_fkey", Def: "FOREIGN KEY (\"markedByEmployeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "VipCustomer_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"VipCustomer_markedByEmployeeId_idx\" ON public.\"VipCustomer\" USING btree (\"markedByEmployeeId\")",
		},
	})
}
