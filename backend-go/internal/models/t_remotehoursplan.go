// موديل جدول "RemoteHoursPlan" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "RemoteHoursPlan",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "targetHours", Type: "integer", NotNull: true},
			{Name: "hourlyRateIQD", Type: "numeric", NotNull: true, Default: "0"},
			{Name: "setById", Type: "text"},
			{Name: "setAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "RemoteHoursPlan_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "RemoteHoursPlan_hourlyRateIQD_check", Def: "CHECK ((\"hourlyRateIQD\" >= (0)::numeric))"},
			{Name: "RemoteHoursPlan_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
			{Name: "RemoteHoursPlan_setById_fkey", Def: "FOREIGN KEY (\"setById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "RemoteHoursPlan_targetHours_check", Def: "CHECK (((\"targetHours\" > 0) AND (\"targetHours\" <= 400)))"},
		},
	})
}
