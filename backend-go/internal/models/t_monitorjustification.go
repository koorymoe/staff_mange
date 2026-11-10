// موديل جدول "MonitorJustification" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MonitorJustification",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "reviewId", Type: "text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "monitorId", Type: "text"},
			{Name: "question", Type: "text", NotNull: true},
			{Name: "answer", Type: "text"},
			{Name: "answeredAt", Type: "timestamp without time zone"},
			{Name: "verdict", Type: "text"},
			{Name: "verdictNote", Type: "text"},
			{Name: "decidedAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "projectId", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "MonitorJustification_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "MonitorJustification_monitorId_fkey", Def: "FOREIGN KEY (\"monitorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "MonitorJustification_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MonitorJustification_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX monitor_justification_emp ON public.\"MonitorJustification\" USING btree (\"employeeId\", \"createdAt\" DESC)",
		},
	})
}
