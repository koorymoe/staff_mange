// موديل جدول "QualityIssue" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "QualityIssue",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "category", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "description", Type: "text"},
			{Name: "responsibleEmployeeId", Type: "text"},
			{Name: "reportedById", Type: "text"},
			{Name: "bookingId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "resolvedAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "QualityIssue_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "QualityIssue_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "QualityIssue_reportedById_fkey", Def: "FOREIGN KEY (\"reportedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "QualityIssue_responsibleEmployeeId_fkey", Def: "FOREIGN KEY (\"responsibleEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"QualityIssue_status_idx\" ON public.\"QualityIssue\" USING btree (status)",
		},
	})
}
