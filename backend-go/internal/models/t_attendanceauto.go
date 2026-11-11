// موديل جدول "AttendanceAuto" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AttendanceAuto",
		Columns: []Column{
			{Name: "attendanceId", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AttendanceAuto_attendanceId_fkey", Def: "FOREIGN KEY (\"attendanceId\") REFERENCES \"Attendance\"(id) ON DELETE CASCADE"},
			{Name: "AttendanceAuto_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "AttendanceAuto_pkey", Def: "PRIMARY KEY (\"attendanceId\")"},
		},
		Indexes: []string{
			"CREATE INDEX \"AttendanceAuto_emp_idx\" ON public.\"AttendanceAuto\" USING btree (\"employeeId\", \"createdAt\")",
		},
	})
}
