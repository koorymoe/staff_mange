// موديل جدول "MediaBrief" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MediaBrief",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "projectId", Type: "text", NotNull: true},
			{Name: "createdById", Type: "text"},
			{Name: "engineerId", Type: "text"},
			{Name: "startAt", Type: "timestamp with time zone"},
			{Name: "expectedEndAt", Type: "timestamp with time zone"},
			{Name: "duration", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'NEW'::text"},
			{Name: "shootAt", Type: "timestamp with time zone"},
			{Name: "mediaEmployeeId", Type: "text"},
			{Name: "publishedUrl", Type: "text"},
			{Name: "mediaNotes", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "MediaBrief_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "MediaBrief_engineerId_fkey", Def: "FOREIGN KEY (\"engineerId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "MediaBrief_mediaEmployeeId_fkey", Def: "FOREIGN KEY (\"mediaEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "MediaBrief_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MediaBrief_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE CASCADE"},
			{Name: "MediaBrief_status_check", Def: "CHECK ((status = ANY (ARRAY['NEW'::text, 'SCHEDULED'::text, 'SHOT'::text, 'PUBLISHED'::text, 'CANCELLED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"MediaBrief_project_idx\" ON public.\"MediaBrief\" USING btree (\"projectId\")",
			"CREATE INDEX \"MediaBrief_status_idx\" ON public.\"MediaBrief\" USING btree (status, \"createdAt\" DESC)",
		},
	})
}
