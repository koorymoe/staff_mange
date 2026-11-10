// موديل جدول "EntityAvatarModel" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EntityAvatarModel",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "label", Type: "text", NotNull: true},
			{Name: "fileKey", Type: "text", NotNull: true},
			{Name: "fileType", Type: "text", NotNull: true},
			{Name: "sizeBytes", Type: "bigint", NotNull: true, Default: "0"},
			{Name: "uploadedById", Type: "text"},
			{Name: "archivedAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "isActive", Type: "boolean", NotNull: true, Default: "false"},
		},
		Constraints: []Constraint{
			{Name: "EntityAvatarModel_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "EntityAvatarModel_uploadedById_fkey", Def: "FOREIGN KEY (\"uploadedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"EntityAvatarModel_active_idx\" ON public.\"EntityAvatarModel\" USING btree (\"archivedAt\", \"createdAt\" DESC)",
			"CREATE UNIQUE INDEX \"EntityAvatarModel_one_active_idx\" ON public.\"EntityAvatarModel\" USING btree (\"isActive\") WHERE \"isActive\"",
		},
	})
}
