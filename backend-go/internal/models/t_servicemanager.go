// موديل جدول "ServiceManager" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ServiceManager",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "ServiceManager_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "ServiceManager_employeeId_serviceId_key", Def: "UNIQUE (\"employeeId\", \"serviceId\")"},
			{Name: "ServiceManager_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ServiceManager_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"ServiceManager_employeeId_idx\" ON public.\"ServiceManager\" USING btree (\"employeeId\")",
			"CREATE INDEX \"ServiceManager_serviceId_idx\" ON public.\"ServiceManager\" USING btree (\"serviceId\")",
		},
	})
}
