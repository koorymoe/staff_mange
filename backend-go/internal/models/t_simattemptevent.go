// موديل جدول "SimAttemptEvent" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimAttemptEvent",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "attemptId", Type: "text", NotNull: true},
			{Name: "stepIndex", Type: "integer"},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "payload", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "atMs", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimAttemptEvent_attemptId_fkey", Def: "FOREIGN KEY (\"attemptId\") REFERENCES \"SimAttempt\"(id) ON DELETE CASCADE"},
			{Name: "SimAttemptEvent_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SimAttemptEvent_att_idx\" ON public.\"SimAttemptEvent\" USING btree (\"attemptId\", \"atMs\")",
		},
	})
}
