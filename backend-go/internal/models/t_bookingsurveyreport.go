// موديل جدول "BookingSurveyReport" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingSurveyReport",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "customerWants", Type: "text", NotNull: true},
			{Name: "siteAreaSqm", Type: "double precision"},
			{Name: "siteDetails", Type: "text"},
			{Name: "otherNotes", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "BookingSurveyReport_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingSurveyReport_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "BookingSurveyReport_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"BookingSurveyReport_bookingId_idx\" ON public.\"BookingSurveyReport\" USING btree (\"bookingId\")",
		},
	})
}
