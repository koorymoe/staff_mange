// موديل جدول "CustomerSignature" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CustomerSignature",
		Columns: []Column{
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "signerName", Type: "text", NotNull: true},
			{Name: "signatureImage", Type: "text", NotNull: true},
			{Name: "collectedById", Type: "text"},
			{Name: "signedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "CustomerSignature_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "CustomerSignature_collectedById_fkey", Def: "FOREIGN KEY (\"collectedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "CustomerSignature_pkey", Def: "PRIMARY KEY (\"bookingId\")"},
		},
	})
}
