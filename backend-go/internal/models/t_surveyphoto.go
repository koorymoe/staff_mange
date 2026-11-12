// موديل جدول "SurveyPhoto" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SurveyPhoto",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "projectId", Type: "text"},
			{Name: "fileKey", Type: "text", NotNull: true},
			{Name: "contentType", Type: "text", NotNull: true},
			{Name: "fileName", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "uploadedById", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SurveyPhoto_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "SurveyPhoto_one_owner", Def: "CHECK (((\"bookingId\" IS NULL) <> (\"projectId\" IS NULL)))"},
			{Name: "SurveyPhoto_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SurveyPhoto_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE CASCADE"},
			{Name: "SurveyPhoto_uploadedById_fkey", Def: "FOREIGN KEY (\"uploadedById\") REFERENCES \"Employee\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SurveyPhoto_bookingId_idx\" ON public.\"SurveyPhoto\" USING btree (\"bookingId\")",
			"CREATE INDEX \"SurveyPhoto_projectId_idx\" ON public.\"SurveyPhoto\" USING btree (\"projectId\")",
		},
	})
}
