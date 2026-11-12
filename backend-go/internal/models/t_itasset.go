// موديل جدول "ItAsset" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ItAsset",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true, Default: "'COMPUTER'::text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'ACTIVE'::text"},
			{Name: "brand", Type: "text"},
			{Name: "model", Type: "text"},
			{Name: "serialNumber", Type: "text"},
			{Name: "ipAddress", Type: "text"},
			{Name: "location", Type: "text"},
			{Name: "assignedEmployeeId", Type: "text"},
			{Name: "purchaseDate", Type: "date"},
			{Name: "warrantyUntil", Type: "date"},
			{Name: "notes", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ItAsset_assignedEmployeeId_fkey", Def: "FOREIGN KEY (\"assignedEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ItAsset_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ItAsset_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ItAsset_kind_idx\" ON public.\"ItAsset\" USING btree (kind)",
			"CREATE INDEX \"ItAsset_status_idx\" ON public.\"ItAsset\" USING btree (status)",
		},
	})
}
