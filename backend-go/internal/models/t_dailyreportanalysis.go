// موديل جدول "DailyReportAnalysis" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DailyReportAnalysis",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "reportId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "day", Type: "date", NotNull: true},
			{Name: "verdict", Type: "text", NotNull: true},
			{Name: "summary", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "claims", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "missing", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "notes", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "bodyHash", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DailyReportAnalysis_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "DailyReportAnalysis_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DailyReportAnalysis_reportId_fkey", Def: "FOREIGN KEY (\"reportId\") REFERENCES \"DailyWorkReport\"(id) ON DELETE CASCADE"},
			{Name: "DailyReportAnalysis_reportId_key", Def: "UNIQUE (\"reportId\")"},
			{Name: "DailyReportAnalysis_verdict_check", Def: "CHECK ((verdict = ANY (ARRAY['CONSISTENT'::text, 'GAPS'::text, 'CONTRADICTS'::text, 'NO_EVIDENCE'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"DailyReportAnalysis_day_idx\" ON public.\"DailyReportAnalysis\" USING btree (day)",
		},
	})
}
