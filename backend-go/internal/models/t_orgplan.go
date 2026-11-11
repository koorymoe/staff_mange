// موديل جدول "OrgPlan" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgPlan",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "divisionId", Type: "text", NotNull: true},
			{Name: "period", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "unitId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'DRAFT'::text"},
			{Name: "submittedAt", Type: "timestamp with time zone"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
			{Name: "reviewNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "OrgPlan_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "OrgPlan_divisionId_fkey", Def: "FOREIGN KEY (\"divisionId\") REFERENCES \"OrgDivision\"(id)"},
			{Name: "OrgPlan_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "OrgPlan_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "OrgPlan_status_check", Def: "CHECK ((status = ANY (ARRAY['DRAFT'::text, 'SUBMITTED'::text, 'APPROVED'::text, 'RETURNED'::text])))"},
			{Name: "OrgPlan_unitId_fkey", Def: "FOREIGN KEY (\"unitId\") REFERENCES \"OrgUnit\"(id)"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"OrgPlan_scope_uq\" ON public.\"OrgPlan\" USING btree (\"divisionId\", COALESCE(\"unitId\", ''::text), period)",
		},
	})
}
