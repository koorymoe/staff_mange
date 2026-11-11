// موديل جدول "PerformanceReview" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PerformanceReview",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "evaluatorId", Type: "text", NotNull: true},
			{Name: "rating", Type: "text", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "bookingId", Type: "text"},
			{Name: "commitmentScore", Type: "smallint"},
			{Name: "speedScore", Type: "smallint"},
			{Name: "qualityScore", Type: "smallint"},
		},
		Constraints: []Constraint{
			{Name: "PerformanceReview_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "PerformanceReview_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "PerformanceReview_evaluatorId_fkey", Def: "FOREIGN KEY (\"evaluatorId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "PerformanceReview_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "performance_review_rating_valid", Def: "CHECK ((rating = ANY (ARRAY['POSITIVE'::text, 'NEEDS_TRAINING'::text, 'MISCONDUCT'::text, 'COMMITMENT'::text])))"},
			{Name: "performance_review_scores_range", Def: "CHECK ((((\"commitmentScore\" IS NULL) OR ((\"commitmentScore\" >= 1) AND (\"commitmentScore\" <= 5))) AND ((\"speedScore\" IS NULL) OR ((\"speedScore\" >= 1) AND (\"speedScore\" <= 5))) AND ((\"qualityScore\" IS NULL) OR ((\"qualityScore\" >= 1) AND (\"qualityScore\" <= 5)))))"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"PerformanceReview_booking_employee_uniq\" ON public.\"PerformanceReview\" USING btree (\"bookingId\", \"employeeId\") WHERE (\"bookingId\" IS NOT NULL)",
			"CREATE INDEX \"PerformanceReview_booking_idx\" ON public.\"PerformanceReview\" USING btree (\"bookingId\")",
			"CREATE INDEX \"PerformanceReview_employeeId_idx\" ON public.\"PerformanceReview\" USING btree (\"employeeId\")",
		},
	})
}
