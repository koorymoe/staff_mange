// موديل جدول "Service" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Service",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "category", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "division", Type: "text", NotNull: true, Default: "'ENGINEERING'::text"},
			{Name: "requiresDeviceInfo", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "managerHandlesPaperwork", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "serviceKind", Type: "text"},
			{Name: "slaHours", Type: "integer", NotNull: true, Default: "48"},
			{Name: "minPhotos", Type: "integer", NotNull: true, Default: "0"},
		},
		Constraints: []Constraint{
			{Name: "Service_name_key", Def: "UNIQUE (name)"},
			{Name: "Service_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Service_serviceKind_idx\" ON public.\"Service\" USING btree (\"serviceKind\") WHERE (\"serviceKind\" IS NOT NULL)",
		},
	})
}
