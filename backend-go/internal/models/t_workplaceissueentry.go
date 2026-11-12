// موديل جدول "WorkplaceIssueEntry" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "WorkplaceIssueEntry",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "issueId", Type: "text", NotNull: true},
			{Name: "authorId", Type: "text"},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "WorkplaceIssueEntry_authorId_fkey", Def: "FOREIGN KEY (\"authorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "WorkplaceIssueEntry_issueId_fkey", Def: "FOREIGN KEY (\"issueId\") REFERENCES \"WorkplaceIssue\"(id) ON DELETE CASCADE"},
			{Name: "WorkplaceIssueEntry_kind_check", Def: "CHECK ((kind = ANY (ARRAY['STATEMENT'::text, 'NOTE'::text, 'DECISION'::text, 'FOLLOWUP'::text, 'MATRIX'::text])))"},
			{Name: "WorkplaceIssueEntry_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"WorkplaceIssueEntry_issue_idx\" ON public.\"WorkplaceIssueEntry\" USING btree (\"issueId\", \"createdAt\")",
		},
	})
}
