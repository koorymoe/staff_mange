// موديل جدول "DowntimePeriod" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-24 دازاي: «فترات توقف — مناسبات يتوقف الشغل أو يقل، الموظف ما يتحاسب على الشغل القليل».
func init() {
	register(Table{
		Name: "DowntimePeriod",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "startDate", Type: "date", NotNull: true},
			{Name: "endDate", Type: "date", NotNull: true},
			{Name: "reason", Type: "text", NotNull: true},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DowntimePeriod_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "DowntimePeriod_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "downtime_period_range_ok", Def: "CHECK ((\"endDate\" >= \"startDate\"))"},
		},
	})
}
