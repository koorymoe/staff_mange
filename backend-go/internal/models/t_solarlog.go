// موديل جدول "SolarLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SolarLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "details", Type: "text", NotNull: true},
			{Name: "payload", Type: "jsonb"},
			{Name: "employeeId", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SolarLog_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "SolarLog_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SolarLog_createdAt_idx\" ON public.\"SolarLog\" USING btree (\"createdAt\" DESC)",
		},
	})
}
