// موديل جدول "OrgPlanItem" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "OrgPlanItem",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "planId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "issueId", Type: "text"},
			{Name: "metric", Type: "text"},
			{Name: "baseline", Type: "numeric"},
			{Name: "target", Type: "numeric"},
			{Name: "manualProgress", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "doneAt", Type: "timestamp with time zone"},
			{Name: "dueDate", Type: "date"},
			{Name: "ownerId", Type: "text"},
			{Name: "taskId", Type: "text"},
			{Name: "byMatrix", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "reason", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "OrgPlanItem_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "OrgPlanItem_issueId_fkey", Def: "FOREIGN KEY (\"issueId\") REFERENCES \"OrgIssueReport\"(id) ON DELETE SET NULL"},
			{Name: "OrgPlanItem_kind_check", Def: "CHECK ((kind = ANY (ARRAY['PROBLEM'::text, 'DEVELOPMENT'::text])))"},
			{Name: "OrgPlanItem_manualProgress_check", Def: "CHECK (((\"manualProgress\" >= 0) AND (\"manualProgress\" <= 100)))"},
			{Name: "OrgPlanItem_ownerId_fkey", Def: "FOREIGN KEY (\"ownerId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "OrgPlanItem_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "OrgPlanItem_planId_fkey", Def: "FOREIGN KEY (\"planId\") REFERENCES \"OrgPlan\"(id) ON DELETE CASCADE"},
			{Name: "OrgPlanItem_taskId_fkey", Def: "FOREIGN KEY (\"taskId\") REFERENCES \"ExtraTask\"(id) ON DELETE SET NULL"},
		},
	})
}
