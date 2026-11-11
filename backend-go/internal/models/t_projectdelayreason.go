// موديل جدول "ProjectDelayReason" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectDelayReason",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "projectId", Type: "text", NotNull: true},
			{Name: "stage", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ProjectDelayReason_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ProjectDelayReason_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProjectDelayReason_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"ProjectDelayReason_project\" ON public.\"ProjectDelayReason\" USING btree (\"projectId\", \"createdAt\")",
		},
	})
}
