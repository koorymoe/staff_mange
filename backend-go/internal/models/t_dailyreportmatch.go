// موديل جدول "DailyReportMatch" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DailyReportMatch",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "day", Type: "date", NotNull: true},
			{Name: "reportId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true},
			{Name: "note", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "detail", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "reviewVerdict", Type: "text"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewNote", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DailyReportMatch_employeeId_day_key", Def: "UNIQUE (\"employeeId\", day)"},
			{Name: "DailyReportMatch_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "DailyReportMatch_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DailyReportMatch_reportId_fkey", Def: "FOREIGN KEY (\"reportId\") REFERENCES \"DailyWorkReport\"(id) ON DELETE CASCADE"},
			{Name: "DailyReportMatch_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"DailyReportMatch_day_idx\" ON public.\"DailyReportMatch\" USING btree (day)",
		},
	})
}
