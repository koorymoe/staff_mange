// موديل جدول "QualityFollowUp" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "QualityFollowUp",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "contactNotes", Type: "text"},
			{Name: "contactedByEmployeeId", Type: "text"},
			{Name: "contactedAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "reportType", Type: "text"},
			{Name: "inspectionStatus", Type: "text", NotNull: true, Default: "'NONE'::text"},
			{Name: "inspectionResult", Type: "text"},
			{Name: "inspectionNotes", Type: "text"},
			{Name: "inspectedById", Type: "text"},
			{Name: "inspectedAt", Type: "timestamp with time zone"},
			{Name: "penalizedEmployeeId", Type: "text"},
			{Name: "kpiEvaluationId", Type: "text"},
			{Name: "linkedBookingId", Type: "text"},
			{Name: "monitorVerdict", Type: "text"},
			{Name: "monitorNote", Type: "text"},
			{Name: "monitorById", Type: "text"},
			{Name: "monitorAt", Type: "timestamp without time zone"},
			{Name: "monitorTargetId", Type: "text"},
			{Name: "customerRating", Type: "integer"},
			{Name: "crewRating", Type: "integer"},
			{Name: "contactedByAgent", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "QualityFollowUp_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "QualityFollowUp_contactedByEmployeeId_fkey", Def: "FOREIGN KEY (\"contactedByEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "QualityFollowUp_crewRating_check", Def: "CHECK (((\"crewRating\" >= 1) AND (\"crewRating\" <= 5)))"},
			{Name: "QualityFollowUp_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE CASCADE"},
			{Name: "QualityFollowUp_customerRating_check", Def: "CHECK (((\"customerRating\" >= 1) AND (\"customerRating\" <= 5)))"},
			{Name: "QualityFollowUp_inspectedById_fkey", Def: "FOREIGN KEY (\"inspectedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "QualityFollowUp_linkedBookingId_fkey", Def: "FOREIGN KEY (\"linkedBookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "QualityFollowUp_monitorById_fkey", Def: "FOREIGN KEY (\"monitorById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "QualityFollowUp_monitorTargetId_fkey", Def: "FOREIGN KEY (\"monitorTargetId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "QualityFollowUp_penalizedEmployeeId_fkey", Def: "FOREIGN KEY (\"penalizedEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "QualityFollowUp_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"QualityFollowUp_bookingId_key\" ON public.\"QualityFollowUp\" USING btree (\"bookingId\")",
			"CREATE INDEX \"QualityFollowUp_inspection_idx\" ON public.\"QualityFollowUp\" USING btree (\"inspectionStatus\") WHERE (\"inspectionStatus\" = 'PENDING'::text)",
			"CREATE INDEX \"QualityFollowUp_linkedBookingId_idx\" ON public.\"QualityFollowUp\" USING btree (\"linkedBookingId\") WHERE (\"linkedBookingId\" IS NOT NULL)",
			"CREATE INDEX \"QualityFollowUp_status_idx\" ON public.\"QualityFollowUp\" USING btree (status)",
		},
	})
}
