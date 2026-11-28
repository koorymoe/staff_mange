// موديل جدول "ActionPlan" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.
// (ك) 11-28 زينا: «معالجة الحالات الحرجة» — خطة لموظف عنده مشكلة بالشغل: الهدف، المدة، والنتيجة والإجراء التالي.

package models

func init() {
	register(Table{
		Name: "ActionPlan",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "problem", Type: "text", NotNull: true},
			{Name: "details", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "source", Type: "text", NotNull: true, Default: "'MANAGER'::text"},
			{Name: "target", Type: "text", NotNull: true},
			{Name: "steps", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "startDate", Type: "date", NotNull: true, Default: "CURRENT_DATE"},
			{Name: "dueDate", Type: "date", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "outcome", Type: "text"},
			{Name: "nextAction", Type: "text"},
			{Name: "createdById", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "closedById", Type: "text"},
			{Name: "closedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "ActionPlan_closedById_fkey", Def: "FOREIGN KEY (\"closedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ActionPlan_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE RESTRICT"},
			{Name: "ActionPlan_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "ActionPlan_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ActionPlan_problem_check", Def: "CHECK ((problem = ANY (ARRAY['SLOW'::text, 'ACCURACY'::text, 'TIME_WASTE'::text, 'BEHAVIOR'::text, 'OTHER'::text])))"},
			{Name: "ActionPlan_source_check", Def: "CHECK ((source = ANY (ARRAY['MATRIX'::text, 'MANAGER'::text, 'OWNER'::text])))"},
			{Name: "ActionPlan_status_check", Def: "CHECK ((status = ANY (ARRAY['OPEN'::text, 'RESOLVED'::text, 'NOT_RESOLVED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"ActionPlan_employee_idx\" ON public.\"ActionPlan\" USING btree (\"employeeId\", \"createdAt\")",
			"CREATE INDEX \"ActionPlan_status_idx\" ON public.\"ActionPlan\" USING btree (status, \"dueDate\")",
		},
	})
}
