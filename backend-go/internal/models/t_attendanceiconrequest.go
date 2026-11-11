// موديل جدول "AttendanceIconRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AttendanceIconRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "requestedIcon", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "resolvedAt", Type: "timestamp without time zone"},
			{Name: "resolvedById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "AttendanceIconRequest_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "AttendanceIconRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "AttendanceIconRequest_resolvedById_fkey", Def: "FOREIGN KEY (\"resolvedById\") REFERENCES \"Employee\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AttendanceIconRequest_status_idx\" ON public.\"AttendanceIconRequest\" USING btree (status)",
		},
	})
}
