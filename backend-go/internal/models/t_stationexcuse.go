// موديل جدول "StationExcuse" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.
// (دازاي 11-19) «النظام يعذر وما يعذر»: الموظف يعتذر عن محطة فاتته، ومدير إدارته يقبل أو يرفض.

package models

func init() {
	register(Table{
		Name: "StationExcuse",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "stationKey", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "StationExcuse_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "StationExcuse_bookingId_stationKey_employeeId_key", Def: "UNIQUE (\"bookingId\", \"stationKey\", \"employeeId\")"},
			{Name: "StationExcuse_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "StationExcuse_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
		},
	})
}
