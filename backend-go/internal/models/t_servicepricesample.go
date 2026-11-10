// موديل جدول "ServicePriceSample" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ServicePriceSample",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "serviceId", Type: "text", NotNull: true},
			{Name: "invoiceId", Type: "text"},
			{Name: "amount", Type: "double precision", NotNull: true},
			{Name: "takenById", Type: "text"},
			{Name: "excluded", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "excludeReason", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ServicePriceSample_invoiceId_fkey", Def: "FOREIGN KEY (\"invoiceId\") REFERENCES \"LeaderInvoice\"(id) ON DELETE SET NULL"},
			{Name: "ServicePriceSample_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ServicePriceSample_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE CASCADE"},
			{Name: "ServicePriceSample_takenById_fkey", Def: "FOREIGN KEY (\"takenById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"ServicePriceSample_service_idx\" ON public.\"ServicePriceSample\" USING btree (\"serviceId\", excluded)",
		},
	})
}
