// موديل جدول "DutySnapshot" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DutySnapshot",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "day", Type: "date", NotNull: true},
			{Name: "position", Type: "text", NotNull: true},
			{Name: "doneCount", Type: "integer", NotNull: true},
			{Name: "total", Type: "integer", NotNull: true},
			{Name: "verdict", Type: "text", NotNull: true},
			{Name: "off", Type: "text"},
			{Name: "duties", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DutySnapshot_employeeId_day_key", Def: "UNIQUE (\"employeeId\", day)"},
			{Name: "DutySnapshot_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "DutySnapshot_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"DutySnapshot_day_idx\" ON public.\"DutySnapshot\" USING btree (day)",
		},
	})
}
