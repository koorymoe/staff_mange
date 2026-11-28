// موديل جدول "ActionPlanNote" — متابعات الخطة (كل ملاحظة بتاريخها ومنو كتبها).

package models

func init() {
	register(Table{
		Name: "ActionPlanNote",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "planId", Type: "text", NotNull: true},
			{Name: "note", Type: "text", NotNull: true},
			{Name: "byId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ActionPlanNote_byId_fkey", Def: "FOREIGN KEY (\"byId\") REFERENCES \"Employee\"(id) ON DELETE RESTRICT"},
			{Name: "ActionPlanNote_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ActionPlanNote_planId_fkey", Def: "FOREIGN KEY (\"planId\") REFERENCES \"ActionPlan\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"ActionPlanNote_plan_idx\" ON public.\"ActionPlanNote\" USING btree (\"planId\", \"createdAt\")",
		},
	})
}
