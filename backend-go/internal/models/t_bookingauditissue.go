// موديل جدول "BookingAuditIssue" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingAuditIssue",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "expectedAmount", Type: "double precision"},
			{Name: "actualAmount", Type: "double precision"},
			{Name: "raisedById", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "resolvedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "resolvedById", Type: "text"},
			{Name: "resolvedByName", Type: "text"},
			{Name: "resolveReason", Type: "text"},
			{Name: "actionKind", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "BookingAuditIssue_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingAuditIssue_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "BookingAuditIssue_raisedById_fkey", Def: "FOREIGN KEY (\"raisedById\") REFERENCES \"Employee\"(id)"},
			{Name: "BookingAuditIssue_resolvedById_fkey", Def: "FOREIGN KEY (\"resolvedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"BookingAuditIssue_open_idx\" ON public.\"BookingAuditIssue\" USING btree (status, kind, \"createdAt\" DESC)",
			"CREATE INDEX \"BookingAuditIssue_status_idx\" ON public.\"BookingAuditIssue\" USING btree (status, \"createdAt\" DESC)",
		},
	})
}
