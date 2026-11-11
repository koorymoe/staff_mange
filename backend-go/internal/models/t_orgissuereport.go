// موديل جدول "OrgIssueReport" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgIssueReport",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "divisionId", Type: "text", NotNull: true},
			{Name: "unitId", Type: "text"},
			{Name: "authorId", Type: "text"},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "problem", Type: "text", NotNull: true},
			{Name: "impact", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "scope", Type: "text", NotNull: true, Default: "'SPECIFIC'::text"},
			{Name: "principle", Type: "integer"},
			{Name: "severity", Type: "text", NotNull: true, Default: "'MEDIUM'::text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "resolution", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "resolvedAt", Type: "timestamp with time zone"},
			{Name: "solutionPlan", Type: "text"},
			{Name: "source", Type: "text", NotNull: true, Default: "'HUMAN'::text"},
			{Name: "signal", Type: "text"},
			{Name: "lastSeenAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "OrgIssueReport_authorId_fkey", Def: "FOREIGN KEY (\"authorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "OrgIssueReport_divisionId_fkey", Def: "FOREIGN KEY (\"divisionId\") REFERENCES \"OrgDivision\"(id)"},
			{Name: "OrgIssueReport_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "OrgIssueReport_principle_check", Def: "CHECK (((principle >= 1) AND (principle <= 40)))"},
			{Name: "OrgIssueReport_scope_check", Def: "CHECK ((scope = ANY (ARRAY['SPECIFIC'::text, 'GENERAL'::text])))"},
			{Name: "OrgIssueReport_severity_check", Def: "CHECK ((severity = ANY (ARRAY['LOW'::text, 'MEDIUM'::text, 'HIGH'::text])))"},
			{Name: "OrgIssueReport_status_check", Def: "CHECK ((status = ANY (ARRAY['OPEN'::text, 'IN_PROGRESS'::text, 'RESOLVED'::text])))"},
			{Name: "OrgIssueReport_unitId_fkey", Def: "FOREIGN KEY (\"unitId\") REFERENCES \"OrgUnit\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"OrgIssueReport_div\" ON public.\"OrgIssueReport\" USING btree (\"divisionId\", status, \"createdAt\" DESC)",
			"CREATE UNIQUE INDEX \"OrgIssueReport_signal_open_uq\" ON public.\"OrgIssueReport\" USING btree (signal) WHERE ((signal IS NOT NULL) AND (status <> 'RESOLVED'::text))",
		},
	})
}
