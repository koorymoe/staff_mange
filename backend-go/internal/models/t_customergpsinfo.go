// موديل جدول "CustomerGpsInfo" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CustomerGpsInfo",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "gpsNumber", Type: "text"},
			{Name: "deviceId", Type: "text"},
			{Name: "subscriptionEnd", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "CustomerGpsInfo_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE CASCADE"},
			{Name: "CustomerGpsInfo_customerId_key", Def: "UNIQUE (\"customerId\")"},
			{Name: "CustomerGpsInfo_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
