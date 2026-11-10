// موديل جدول "OrgDirective" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgDirective",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "scope", Type: "text", NotNull: true},
			{Name: "divisionId", Type: "text"},
			{Name: "unitId", Type: "text"},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "body", Type: "text", NotNull: true},
			{Name: "clause", Type: "integer"},
			{Name: "issuedById", Type: "text"},
			{Name: "issuedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "archivedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "OrgDirective_divisionId_fkey", Def: "FOREIGN KEY (\"divisionId\") REFERENCES \"OrgDivision\"(id)"},
			{Name: "OrgDirective_issuedById_fkey", Def: "FOREIGN KEY (\"issuedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "OrgDirective_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "OrgDirective_scope_check", Def: "CHECK ((scope = ANY (ARRAY['COMPANY'::text, 'DIVISION'::text, 'UNIT'::text])))"},
			{Name: "OrgDirective_unitId_fkey", Def: "FOREIGN KEY (\"unitId\") REFERENCES \"OrgUnit\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"OrgDirective_active\" ON public.\"OrgDirective\" USING btree (\"issuedAt\" DESC) WHERE (\"archivedAt\" IS NULL)",
		},
	})
}
