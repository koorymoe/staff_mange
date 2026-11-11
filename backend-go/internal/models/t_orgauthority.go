// موديل جدول "OrgAuthority" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgAuthority",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "divisionId", Type: "text", NotNull: true},
			{Name: "decision", Type: "text", NotNull: true},
			{Name: "mode", Type: "text", NotNull: true, Default: "'DECIDE'::text"},
			{Name: "limitIQD", Type: "numeric"},
			{Name: "sort", Type: "integer", NotNull: true, Default: "0"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "OrgAuthority_divisionId_fkey", Def: "FOREIGN KEY (\"divisionId\") REFERENCES \"OrgDivision\"(id)"},
			{Name: "OrgAuthority_mode_check", Def: "CHECK ((mode = ANY (ARRAY['DECIDE'::text, 'ESCALATE'::text])))"},
			{Name: "OrgAuthority_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "OrgAuthority_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"OrgAuthority_div\" ON public.\"OrgAuthority\" USING btree (\"divisionId\", sort)",
		},
	})
}
