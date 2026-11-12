// موديل جدول "ProjectDelegationLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectDelegationLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "projectId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "delegatedByEmployeeId", Type: "text"},
			{Name: "action", Type: "text", NotNull: true, Default: "'ASSIGN'::text"},
			{Name: "note", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ProjectDelegationLog_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProjectDelegationLog_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"ProjectDelegationLog_employee_idx\" ON public.\"ProjectDelegationLog\" USING btree (\"employeeId\")",
			"CREATE INDEX \"ProjectDelegationLog_project_idx\" ON public.\"ProjectDelegationLog\" USING btree (\"projectId\")",
		},
	})
}
