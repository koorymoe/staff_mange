// موديل جدول "Skill" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Skill",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "category", Type: "text", NotNull: true, Default: "'فنية'::text"},
			{Name: "description", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "Skill_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "Skill_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
			{Name: "Skill_serviceId_name_key", Def: "UNIQUE (\"serviceId\", name)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Skill_category_idx\" ON public.\"Skill\" USING btree (category)",
			"CREATE INDEX \"Skill_serviceId_idx\" ON public.\"Skill\" USING btree (\"serviceId\")",
		},
	})
}
