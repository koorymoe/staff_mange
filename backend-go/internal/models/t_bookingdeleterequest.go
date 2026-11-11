// موديل جدول "BookingDeleteRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingDeleteRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "requestedById", Type: "text", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "decisionNote", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "channel", Type: "text"},
			{Name: "requestType", Type: "text"},
			{Name: "needsInfo", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "needsInfoNote", Type: "text"},
			{Name: "needsInfoAt", Type: "timestamp without time zone"},
			{Name: "needsInfoById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "BookingDeleteRequest_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingDeleteRequest_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id)"},
			{Name: "BookingDeleteRequest_needsInfoById_fkey", Def: "FOREIGN KEY (\"needsInfoById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "BookingDeleteRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "BookingDeleteRequest_requestedById_fkey", Def: "FOREIGN KEY (\"requestedById\") REFERENCES \"Employee\"(id)"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"BookingDeleteRequest_one_pending\" ON public.\"BookingDeleteRequest\" USING btree (\"bookingId\") WHERE (status = 'PENDING'::text)",
			"CREATE INDEX \"BookingDeleteRequest_status_needsinfo_idx\" ON public.\"BookingDeleteRequest\" USING btree (status, \"needsInfo\", \"createdAt\" DESC)",
		},
	})
}
