// موديل جدول "ProjectMonitorAction" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectMonitorAction",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "projectId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "action", Type: "text", NotNull: true},
			{Name: "note", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "dueAt", Type: "timestamp with time zone"},
			{Name: "points", Type: "integer", NotNull: true, Default: "0"},
			{Name: "taskId", Type: "text"},
			{Name: "issueId", Type: "text"},
			{Name: "byId", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "revokedAt", Type: "timestamp with time zone"},
			{Name: "revokedById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "ProjectMonitorAction_action_check", Def: "CHECK ((action = ANY (ARRAY['FOLLOW_UP'::text, 'ESCALATE'::text, 'VIOLATION'::text])))"},
			{Name: "ProjectMonitorAction_byId_fkey", Def: "FOREIGN KEY (\"byId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ProjectMonitorAction_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ProjectMonitorAction_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProjectMonitorAction_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE CASCADE"},
			{Name: "ProjectMonitorAction_revokedById_fkey", Def: "FOREIGN KEY (\"revokedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"ProjectMonitorAction_project\" ON public.\"ProjectMonitorAction\" USING btree (\"projectId\", \"createdAt\" DESC)",
		},
	})
}
