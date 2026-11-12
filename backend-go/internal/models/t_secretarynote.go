// موديل جدول "SecretaryNote" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SecretaryNote",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "ownerId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "body", Type: "text", NotNull: true},
			{Name: "at", Type: "timestamp with time zone"},
			{Name: "doneAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SecretaryNote_kind_check", Def: "CHECK ((kind = ANY (ARRAY['DIARY'::text, 'AGENDA'::text])))"},
			{Name: "SecretaryNote_ownerId_fkey", Def: "FOREIGN KEY (\"ownerId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "SecretaryNote_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SecretaryNote_owner_idx\" ON public.\"SecretaryNote\" USING btree (\"ownerId\", kind, \"createdAt\" DESC)",
		},
	})
}
