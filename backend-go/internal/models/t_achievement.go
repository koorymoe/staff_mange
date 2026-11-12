// موديل جدول "Achievement" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Achievement",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "reportText", Type: "text", NotNull: true},
			{Name: "reviewStatus", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "reviewNote", Type: "text"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "Achievement_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "Achievement_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "Achievement_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "Achievement_reviewStatus_check", Def: "CHECK ((\"reviewStatus\" = ANY (ARRAY['PENDING'::text, 'GOOD'::text, 'NEEDS_REVIEW'::text])))"},
			{Name: "Achievement_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"Achievement_createdAt_idx\" ON public.\"Achievement\" USING btree (\"createdAt\")",
			"CREATE INDEX \"Achievement_employeeId_idx\" ON public.\"Achievement\" USING btree (\"employeeId\", \"createdAt\" DESC)",
			"CREATE INDEX \"Achievement_reviewStatus_idx\" ON public.\"Achievement\" USING btree (\"reviewStatus\")",
		},
	})
}
