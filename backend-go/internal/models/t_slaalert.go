// موديل جدول "SlaAlert" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SlaAlert",
		Columns: []Column{
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SlaAlert_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "SlaAlert_pkey", Def: "PRIMARY KEY (\"bookingId\", kind)"},
		},
	})
}
