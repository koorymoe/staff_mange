// موديل جدول "Customer" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Customer",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "customerCode", Type: "integer", NotNull: true, Default: "nextval('\"Customer_customerCode_seq\"'::regclass)"},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "phone", Type: "text", NotNull: true},
			{Name: "location", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "mapLatitude", Type: "double precision"},
			{Name: "mapLongitude", Type: "double precision"},
			{Name: "position", Type: "text"},
			{Name: "locationUrl", Type: "text"},
			{Name: "falseClaimCount", Type: "integer", NotNull: true, Default: "0"},
			{Name: "lastFalseClaimAt", Type: "timestamp with time zone"},
			{Name: "falseClaimNote", Type: "text"},
			{Name: "lat", Type: "double precision"},
			{Name: "lng", Type: "double precision"},
			{Name: "locationSource", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "Customer_customerCode_key", Def: "UNIQUE (\"customerCode\")"},
			{Name: "Customer_phone_key", Def: "UNIQUE (phone)"},
			{Name: "Customer_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Customer_name_trgm_idx\" ON public.\"Customer\" USING gin (ar_norm(name) gin_trgm_ops)",
			"CREATE INDEX \"Customer_phone_norm_idx\" ON public.\"Customer\" USING btree (phone_norm(phone))",
			"CREATE INDEX \"Customer_phone_trgm_idx\" ON public.\"Customer\" USING gin (phone gin_trgm_ops)",
		},
	})
}
