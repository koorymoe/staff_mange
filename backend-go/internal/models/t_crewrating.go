// موديل جدول "CrewRating" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CrewRating",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "leaderId", Type: "text", NotNull: true},
			{Name: "technicianId", Type: "text", NotNull: true},
			{Name: "score", Type: "integer", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "CrewRating_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "CrewRating_bookingId_technicianId_key", Def: "UNIQUE (\"bookingId\", \"technicianId\")"},
			{Name: "CrewRating_leaderId_fkey", Def: "FOREIGN KEY (\"leaderId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "CrewRating_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "CrewRating_score_check", Def: "CHECK (((score >= 1) AND (score <= 5)))"},
			{Name: "CrewRating_technicianId_fkey", Def: "FOREIGN KEY (\"technicianId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"CrewRating_technician_idx\" ON public.\"CrewRating\" USING btree (\"technicianId\", \"createdAt\" DESC)",
		},
	})
}
