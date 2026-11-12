// موديل جدول "Complaint" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Complaint",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "description", Type: "text", NotNull: true},
			{Name: "status", Type: "\"ComplaintStatus\"", NotNull: true, Default: "'NEW'::\"ComplaintStatus\""},
			{Name: "createdByEmployeeId", Type: "text", NotNull: true},
			{Name: "assignedToEmployeeId", Type: "text"},
			{Name: "resolution", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "resolvedAt", Type: "timestamp without time zone"},
			{Name: "type", Type: "text", NotNull: true, Default: "'OTHER'::text"},
			{Name: "relatedEmployeeId", Type: "text"},
			{Name: "relatedEmployeeName", Type: "text"},
			{Name: "contactedAt", Type: "timestamp with time zone"},
			{Name: "contactedById", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "customerRating", Type: "integer"},
			{Name: "auditVerdict", Type: "text"},
			{Name: "auditNote", Type: "text"},
			{Name: "auditedAt", Type: "timestamp without time zone"},
			{Name: "auditedById", Type: "text"},
			{Name: "settledLegacyAt", Type: "timestamp without time zone"},
			{Name: "noAnswerCount", Type: "integer", NotNull: true, Default: "0"},
			{Name: "noAnswerAt", Type: "timestamp without time zone"},
			{Name: "nextCallAt", Type: "timestamp without time zone"},
			{Name: "droppedNoAnswerAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "Complaint_assignedToEmployeeId_fkey", Def: "FOREIGN KEY (\"assignedToEmployeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "Complaint_auditedById_fkey", Def: "FOREIGN KEY (\"auditedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "Complaint_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id)"},
			{Name: "Complaint_contactedById_fkey", Def: "FOREIGN KEY (\"contactedById\") REFERENCES \"Employee\"(id)"},
			{Name: "Complaint_createdByEmployeeId_fkey", Def: "FOREIGN KEY (\"createdByEmployeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "Complaint_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id)"},
			{Name: "Complaint_customerRating_range", Def: "CHECK (((\"customerRating\" IS NULL) OR ((\"customerRating\" >= 1) AND (\"customerRating\" <= 5))))"},
			{Name: "Complaint_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "Complaint_relatedEmployeeId_fkey", Def: "FOREIGN KEY (\"relatedEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"Complaint_auditVerdict_idx\" ON public.\"Complaint\" USING btree (\"auditVerdict\")",
			"CREATE INDEX \"Complaint_customerId_idx\" ON public.\"Complaint\" USING btree (\"customerId\")",
			"CREATE INDEX \"Complaint_status_idx\" ON public.\"Complaint\" USING btree (status)",
		},
	})
}
