// موديل جدول "RemoteWorkEntry" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "RemoteWorkEntry",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "workDate", Type: "date", NotNull: true},
			{Name: "seconds", Type: "integer", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "addedById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "closedAt", Type: "timestamp with time zone"},
			{Name: "closedById", Type: "text"},
			{Name: "source", Type: "text", NotNull: true, Default: "'MANUAL'::text"},
		},
		Constraints: []Constraint{
			{Name: "RemoteWorkEntry_addedById_fkey", Def: "FOREIGN KEY (\"addedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "RemoteWorkEntry_closedById_fkey", Def: "FOREIGN KEY (\"closedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "RemoteWorkEntry_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "RemoteWorkEntry_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "RemoteWorkEntry_seconds_check", Def: "CHECK ((seconds > 0))"},
		},
		Indexes: []string{
			"CREATE INDEX \"RemoteWorkEntry_emp_date_idx\" ON public.\"RemoteWorkEntry\" USING btree (\"employeeId\", \"workDate\")",
		},
	})
}
