// موديل جدول "ServiceStandardTime" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ServiceStandardTime",
		Columns: []Column{
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "minutes", Type: "integer", NotNull: true},
			{Name: "setById", Type: "text"},
			{Name: "setAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ServiceStandardTime_minutes_check", Def: "CHECK (((minutes >= 5) AND (minutes <= 1440)))"},
			{Name: "ServiceStandardTime_pkey", Def: "PRIMARY KEY (\"serviceId\")"},
			{Name: "ServiceStandardTime_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
		},
	})
}
