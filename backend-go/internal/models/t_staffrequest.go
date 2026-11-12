// موديل جدول "StaffRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "StaffRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "requesterId", Type: "text", NotNull: true},
			{Name: "projectId", Type: "text"},
			{Name: "neededAt", Type: "timestamp without time zone", NotNull: true},
			{Name: "durationHours", Type: "double precision", NotNull: true, Default: "8"},
			{Name: "notes", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "handledById", Type: "text"},
			{Name: "handledAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "StaffRequest_handledById_fkey", Def: "FOREIGN KEY (\"handledById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "StaffRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "StaffRequest_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE SET NULL"},
			{Name: "StaffRequest_requesterId_fkey", Def: "FOREIGN KEY (\"requesterId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"StaffRequest_requesterId_idx\" ON public.\"StaffRequest\" USING btree (\"requesterId\")",
			"CREATE INDEX \"StaffRequest_status_idx\" ON public.\"StaffRequest\" USING btree (status)",
		},
	})
}
