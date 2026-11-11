// موديل جدول "OrgAssignment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgAssignment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "divisionId", Type: "text", NotNull: true},
			{Name: "unitId", Type: "text"},
			{Name: "position", Type: "text", NotNull: true},
			{Name: "allowanceIQD", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "acting", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "actingUntil", Type: "timestamp with time zone"},
			{Name: "replacesId", Type: "text"},
			{Name: "startedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "endedAt", Type: "timestamp with time zone"},
			{Name: "endedById", Type: "text"},
			{Name: "endReason", Type: "text"},
			{Name: "note", Type: "text"},
			{Name: "createdById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "OrgAssignment_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "OrgAssignment_divisionId_fkey", Def: "FOREIGN KEY (\"divisionId\") REFERENCES \"OrgDivision\"(id)"},
			{Name: "OrgAssignment_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "OrgAssignment_endedById_fkey", Def: "FOREIGN KEY (\"endedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "OrgAssignment_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "OrgAssignment_position_check", Def: "CHECK ((\"position\" = ANY (ARRAY['DIRECTOR'::text, 'DEPUTY'::text, 'MEMBER'::text])))"},
			{Name: "OrgAssignment_replacesId_fkey", Def: "FOREIGN KEY (\"replacesId\") REFERENCES \"OrgAssignment\"(id)"},
			{Name: "OrgAssignment_unitId_fkey", Def: "FOREIGN KEY (\"unitId\") REFERENCES \"OrgUnit\"(id)"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"OrgAssignment_acting_uq\" ON public.\"OrgAssignment\" USING btree (\"replacesId\") WHERE ((\"endedAt\" IS NULL) AND acting)",
			"CREATE INDEX \"OrgAssignment_emp\" ON public.\"OrgAssignment\" USING btree (\"employeeId\") WHERE (\"endedAt\" IS NULL)",
			"CREATE UNIQUE INDEX \"OrgAssignment_emp_uq\" ON public.\"OrgAssignment\" USING btree (\"employeeId\", \"divisionId\", COALESCE(\"unitId\", ''::text)) WHERE ((\"endedAt\" IS NULL) AND (NOT acting))",
			"CREATE UNIQUE INDEX \"OrgAssignment_head_uq\" ON public.\"OrgAssignment\" USING btree (\"divisionId\", COALESCE(\"unitId\", ''::text), \"position\") WHERE ((\"endedAt\" IS NULL) AND (NOT acting) AND (\"position\" = ANY (ARRAY['DIRECTOR'::text, 'DEPUTY'::text])))",
		},
	})
}
