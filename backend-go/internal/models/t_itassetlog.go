// موديل جدول "ItAssetLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ItAssetLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "assetId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true, Default: "'NOTE'::text"},
			{Name: "note", Type: "text", NotNull: true},
			{Name: "cost", Type: "double precision"},
			{Name: "employeeId", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ItAssetLog_assetId_fkey", Def: "FOREIGN KEY (\"assetId\") REFERENCES \"ItAsset\"(id) ON DELETE CASCADE"},
			{Name: "ItAssetLog_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ItAssetLog_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ItAssetLog_assetId_idx\" ON public.\"ItAssetLog\" USING btree (\"assetId\", \"createdAt\" DESC)",
		},
	})
}
