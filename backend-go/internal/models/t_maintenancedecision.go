// موديل جدول "MaintenanceDecision" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.
// (بلاغ #55، 11-20) قرار الجودة على زبون الصيانة الوقائية: حجز / ما يريد / ما رد / أجّل.

package models

func init() {
	register(Table{
		Name: "MaintenanceDecision",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "decision", Type: "text", NotNull: true},
			{Name: "note", Type: "text"},
			{Name: "hiddenUntil", Type: "timestamp with time zone", NotNull: true},
			{Name: "byId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "MaintenanceDecision_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MaintenanceDecision_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"MaintenanceDecision_customer_idx\" ON public.\"MaintenanceDecision\" USING btree (\"customerId\", \"hiddenUntil\" DESC)",
		},
	})
}
