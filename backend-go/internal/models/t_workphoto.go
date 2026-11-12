// موديل جدول "WorkPhoto" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "WorkPhoto",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "phase", Type: "text", NotNull: true},
			{Name: "fileKey", Type: "text", NotNull: true},
			{Name: "contentType", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "WorkPhoto_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "WorkPhoto_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "WorkPhoto_phase_check", Def: "CHECK ((phase = ANY (ARRAY['BEFORE'::text, 'AFTER'::text])))"},
			{Name: "WorkPhoto_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"WorkPhoto_booking\" ON public.\"WorkPhoto\" USING btree (\"bookingId\", phase)",
		},
	})
}
