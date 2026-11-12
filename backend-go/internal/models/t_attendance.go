// موديل جدول "Attendance" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Attendance",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "checkIn", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "checkOut", Type: "timestamp without time zone"},
			{Name: "date", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "approvedOut", Type: "boolean", NotNull: true, Default: "false"},
		},
		Constraints: []Constraint{
			{Name: "Attendance_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "Attendance_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Attendance_date_idx\" ON public.\"Attendance\" USING btree (date)",
			"CREATE INDEX \"Attendance_employeeId_idx\" ON public.\"Attendance\" USING btree (\"employeeId\")",
			"CREATE UNIQUE INDEX \"Attendance_one_open_idx\" ON public.\"Attendance\" USING btree (\"employeeId\") WHERE (\"checkOut\" IS NULL)",
		},
	})
}
