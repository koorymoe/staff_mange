// موديل جدول "EmployeeIdea" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.
// (دازاي 11-21، أفكار ٤ — كايزن) «💡 عندي فكرة»: الموظف يقترح تحسين لشغله، ومسؤول إدارته يرد خلال ٤٨ ساعة.

package models

func init() {
	register(Table{
		Name: "EmployeeIdea",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "text", Type: "text", NotNull: true},
			{Name: "category", Type: "text", NotNull: true, Default: "'OTHER'::text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'NEW'::text"},
			{Name: "reply", Type: "text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "escalatedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeIdea_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "EmployeeIdea_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"EmployeeIdea_status_idx\" ON public.\"EmployeeIdea\" USING btree (status, \"createdAt\" DESC)",
		},
	})
}
