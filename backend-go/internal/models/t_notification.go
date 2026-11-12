// موديل جدول "Notification" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Notification",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "type", Type: "text", NotNull: true},
			{Name: "message", Type: "text", NotNull: true},
			{Name: "read", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "pushedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "Notification_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "Notification_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Notification_employeeId_idx\" ON public.\"Notification\" USING btree (\"employeeId\", \"createdAt\" DESC)",
			"CREATE INDEX \"Notification_push_pending_idx\" ON public.\"Notification\" USING btree (\"createdAt\") WHERE (\"pushedAt\" IS NULL)",
			"CREATE INDEX \"Notification_unread\" ON public.\"Notification\" USING btree (\"employeeId\") WHERE (read = false)",
		},
	})
}
