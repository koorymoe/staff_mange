// موديل جدول "OrgComment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgComment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "targetType", Type: "text", NotNull: true},
			{Name: "targetId", Type: "text", NotNull: true},
			{Name: "authorId", Type: "text"},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "OrgComment_authorId_fkey", Def: "FOREIGN KEY (\"authorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "OrgComment_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "OrgComment_targetType_check", Def: "CHECK ((\"targetType\" = ANY (ARRAY['ISSUE'::text, 'PLAN_ITEM'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"OrgComment_target\" ON public.\"OrgComment\" USING btree (\"targetType\", \"targetId\", \"createdAt\")",
		},
	})
}
