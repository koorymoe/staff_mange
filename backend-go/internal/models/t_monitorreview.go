// موديل جدول "MonitorReview" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MonitorReview",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "stage", Type: "text", NotNull: true},
			{Name: "entityType", Type: "text", NotNull: true},
			{Name: "entityId", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "summary", Type: "text"},
			{Name: "ownerRole", Type: "text"},
			{Name: "ownerEmployeeId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "note", Type: "text"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "urgent", Type: "boolean", NotNull: true, Default: "false"},
		},
		Constraints: []Constraint{
			{Name: "MonitorReview_ownerEmployeeId_fkey", Def: "FOREIGN KEY (\"ownerEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "MonitorReview_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MonitorReview_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"MonitorReview_pending_idx\" ON public.\"MonitorReview\" USING btree (status, \"createdAt\" DESC)",
			"CREATE INDEX \"MonitorReview_stage_idx\" ON public.\"MonitorReview\" USING btree (stage, status)",
			"CREATE UNIQUE INDEX \"MonitorReview_unique_stage_idx\" ON public.\"MonitorReview\" USING btree (\"entityType\", \"entityId\", stage)",
			"CREATE INDEX \"MonitorReview_urgent_idx\" ON public.\"MonitorReview\" USING btree (urgent, \"createdAt\" DESC) WHERE (status = 'PENDING'::text)",
		},
	})
}
