// موديل جدول "WorkplaceIssue" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "WorkplaceIssue",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "partyAId", Type: "text", NotNull: true},
			{Name: "partyBId", Type: "text"},
			{Name: "openedById", Type: "text"},
			{Name: "description", Type: "text", NotNull: true},
			{Name: "severity", Type: "text", NotNull: true, Default: "'ATTENTION'::text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "aiSummary", Type: "text"},
			{Name: "decision", Type: "text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "followUpAt", Type: "timestamp with time zone"},
			{Name: "sourceNoteId", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "WorkplaceIssue_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "WorkplaceIssue_openedById_fkey", Def: "FOREIGN KEY (\"openedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "WorkplaceIssue_partyAId_fkey", Def: "FOREIGN KEY (\"partyAId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "WorkplaceIssue_partyBId_fkey", Def: "FOREIGN KEY (\"partyBId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "WorkplaceIssue_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "WorkplaceIssue_severity_check", Def: "CHECK ((severity = ANY (ARRAY['NORMAL'::text, 'ATTENTION'::text, 'URGENT'::text])))"},
			{Name: "WorkplaceIssue_sourceNoteId_fkey", Def: "FOREIGN KEY (\"sourceNoteId\") REFERENCES \"PeerNote\"(id) ON DELETE SET NULL"},
			{Name: "WorkplaceIssue_status_check", Def: "CHECK ((status = ANY (ARRAY['OPEN'::text, 'IN_PROGRESS'::text, 'RESOLVED'::text])))"},
		},
	})
}
