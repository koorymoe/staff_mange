// موديل جدول "DailyWorkReport" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DailyWorkReport",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "day", Type: "date", NotNull: true},
			{Name: "body", Type: "text", NotNull: true},
			{Name: "late", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DailyWorkReport_body_check", Def: "CHECK ((length(btrim(body)) >= 20))"},
			{Name: "DailyWorkReport_employeeId_day_key", Def: "UNIQUE (\"employeeId\", day)"},
			{Name: "DailyWorkReport_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "DailyWorkReport_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"DailyWorkReport_day\" ON public.\"DailyWorkReport\" USING btree (day DESC)",
		},
	})
}
