// موديل جدول "ManagerRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ManagerRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "requesterId", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "details", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "priority", Type: "text", NotNull: true, Default: "'NORMAL'::text"},
			{Name: "context", Type: "text"},
			{Name: "source", Type: "text", NotNull: true, Default: "'SCREEN'::text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "reply", Type: "text"},
			{Name: "handledById", Type: "text"},
			{Name: "handledAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "remindedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "ManagerRequest_handledById_fkey", Def: "FOREIGN KEY (\"handledById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ManagerRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ManagerRequest_priority_check", Def: "CHECK ((priority = ANY (ARRAY['NORMAL'::text, 'HIGH'::text, 'URGENT'::text])))"},
			{Name: "ManagerRequest_requesterId_fkey", Def: "FOREIGN KEY (\"requesterId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "ManagerRequest_status_check", Def: "CHECK ((status = ANY (ARRAY['OPEN'::text, 'DONE'::text, 'REJECTED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"ManagerRequest_open_idx\" ON public.\"ManagerRequest\" USING btree (status, \"createdAt\")",
			"CREATE INDEX \"ManagerRequest_requester_idx\" ON public.\"ManagerRequest\" USING btree (\"requesterId\", \"createdAt\")",
		},
	})
}
