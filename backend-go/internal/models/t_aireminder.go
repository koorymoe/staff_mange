// موديل جدول "AiReminder" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiReminder",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "dueAt", Type: "timestamp with time zone", NotNull: true},
			{Name: "sentAt", Type: "timestamp with time zone"},
			{Name: "cancelledAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiReminder_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "AiReminder_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AiReminder_due_idx\" ON public.\"AiReminder\" USING btree (\"dueAt\") WHERE ((\"sentAt\" IS NULL) AND (\"cancelledAt\" IS NULL))",
		},
	})
}
