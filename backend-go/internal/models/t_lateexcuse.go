// موديل جدول "LateExcuse" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "LateExcuse",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "date", Type: "date", NotNull: true},
			{Name: "minutes", Type: "integer"},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "decisionNote", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "LateExcuse_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "LateExcuse_employeeId_date_key", Def: "UNIQUE (\"employeeId\", date)"},
			{Name: "LateExcuse_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "LateExcuse_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "LateExcuse_status_check", Def: "CHECK ((status = ANY (ARRAY['PENDING'::text, 'APPROVED'::text, 'REJECTED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"LateExcuse_status_idx\" ON public.\"LateExcuse\" USING btree (status) WHERE (status = 'PENDING'::text)",
		},
	})
}
