// موديل جدول "GpsRenewalFollowUp" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "GpsRenewalFollowUp",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "deviceRequestId", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text"},
			{Name: "calledById", Type: "text"},
			{Name: "outcome", Type: "text", NotNull: true},
			{Name: "notes", Type: "text"},
			{Name: "daysSinceExpiry", Type: "integer"},
			{Name: "calledAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "GpsRenewalFollowUp_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"GpsRenewalFollowUp_device_idx\" ON public.\"GpsRenewalFollowUp\" USING btree (\"deviceRequestId\", \"calledAt\" DESC)",
		},
	})
}
