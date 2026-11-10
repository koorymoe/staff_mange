// موديل جدول "CustomerRating" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CustomerRating",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "accountId", Type: "text", NotNull: true},
			{Name: "stars", Type: "integer", NotNull: true},
			{Name: "comment", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "CustomerRating_accountId_fkey", Def: "FOREIGN KEY (\"accountId\") REFERENCES \"CustomerAccount\"(id) ON DELETE CASCADE"},
			{Name: "CustomerRating_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "CustomerRating_bookingId_key", Def: "UNIQUE (\"bookingId\")"},
			{Name: "CustomerRating_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "CustomerRating_stars_check", Def: "CHECK (((stars >= 1) AND (stars <= 5)))"},
		},
	})
}
