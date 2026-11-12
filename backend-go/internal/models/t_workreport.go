// موديل جدول "WorkReport" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "WorkReport",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "workStatus", Type: "\"WorkStatus\"", NotNull: true},
			{Name: "events", Type: "text"},
			{Name: "extraRequests", Type: "text"},
			{Name: "cleanedSite", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "gaveInfo", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "tookPhotos", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "stopReason", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "sheet", Type: "jsonb"},
			{Name: "reviewedAt", Type: "timestamp without time zone"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "WorkReport_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id)"},
			{Name: "WorkReport_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "WorkReport_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "WorkReport_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"WorkReport_bookingId_idx\" ON public.\"WorkReport\" USING btree (\"bookingId\")",
			"CREATE INDEX \"WorkReport_employeeId_idx\" ON public.\"WorkReport\" USING btree (\"employeeId\")",
		},
	})
}
