// موديل جدول "SimCard" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimCard",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "simNumber", Type: "text", NotNull: true},
			{Name: "iccid", Type: "text"},
			{Name: "operator", Type: "\"SimOperator\"", NotNull: true},
			{Name: "status", Type: "\"SimStatus\"", NotNull: true, Default: "'AVAILABLE'::\"SimStatus\""},
			{Name: "customerId", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "assignedAt", Type: "timestamp with time zone"},
			{Name: "releasedAt", Type: "timestamp with time zone"},
			{Name: "burnedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "SimCard_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"GpsCustomer\"(id)"},
			{Name: "SimCard_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SimCard_simNumber_key", Def: "UNIQUE (\"simNumber\")"},
		},
		Indexes: []string{
			"CREATE INDEX \"SimCard_status_idx\" ON public.\"SimCard\" USING btree (status)",
		},
	})
}
