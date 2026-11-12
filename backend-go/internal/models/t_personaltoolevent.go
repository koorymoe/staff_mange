// موديل جدول "PersonalToolEvent" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PersonalToolEvent",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "toolId", Type: "text", NotNull: true},
			{Name: "toolName", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "eventType", Type: "text", NotNull: true},
			{Name: "fromStatus", Type: "text"},
			{Name: "toStatus", Type: "text"},
			{Name: "note", Type: "text"},
			{Name: "actorId", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "PersonalToolEvent_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"PersonalToolEvent_employeeId_idx\" ON public.\"PersonalToolEvent\" USING btree (\"employeeId\", \"createdAt\" DESC)",
			"CREATE INDEX \"PersonalToolEvent_toolId_idx\" ON public.\"PersonalToolEvent\" USING btree (\"toolId\", \"createdAt\" DESC)",
			"CREATE INDEX \"PersonalToolEvent_type_idx\" ON public.\"PersonalToolEvent\" USING btree (\"eventType\", \"createdAt\" DESC)",
		},
	})
}
