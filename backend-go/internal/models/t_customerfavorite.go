// موديل جدول "CustomerFavorite" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CustomerFavorite",
		Columns: []Column{
			{Name: "accountId", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "CustomerFavorite_accountId_fkey", Def: "FOREIGN KEY (\"accountId\") REFERENCES \"CustomerAccount\"(id) ON DELETE CASCADE"},
			{Name: "CustomerFavorite_pkey", Def: "PRIMARY KEY (\"accountId\", \"serviceId\")"},
			{Name: "CustomerFavorite_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
		},
	})
}
