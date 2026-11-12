// موديل جدول "Announcement" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Announcement",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "body", Type: "text", NotNull: true},
			{Name: "active", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdById", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "expiresAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "Announcement_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "Announcement_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Announcement_active_idx\" ON public.\"Announcement\" USING btree (active, \"createdAt\" DESC)",
			"CREATE INDEX \"Announcement_expiry_idx\" ON public.\"Announcement\" USING btree (active, \"expiresAt\")",
		},
	})
}
