// موديل جدول "GpsSubscriptionPrice" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "GpsSubscriptionPrice",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "subscriptionType", Type: "\"GpsSubscriptionType\"", NotNull: true},
			{Name: "price", Type: "double precision", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "GpsSubscriptionPrice_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "GpsSubscriptionPrice_subscriptionType_key", Def: "UNIQUE (\"subscriptionType\")"},
		},
	})
}
