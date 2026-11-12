// موديل جدول "CustomerServiceTag" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CustomerServiceTag",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "service", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "CustomerServiceTag_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE CASCADE"},
			{Name: "CustomerServiceTag_customerId_service_key", Def: "UNIQUE (\"customerId\", service)"},
			{Name: "CustomerServiceTag_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"CustomerServiceTag_customerId_idx\" ON public.\"CustomerServiceTag\" USING btree (\"customerId\")",
			"CREATE INDEX \"CustomerServiceTag_service_idx\" ON public.\"CustomerServiceTag\" USING btree (service)",
		},
	})
}
