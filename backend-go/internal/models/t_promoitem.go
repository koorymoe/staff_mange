// موديل جدول "PromoItem" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PromoItem",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true, Default: "'offer'::text"},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "body", Type: "text"},
			{Name: "imageUrl", Type: "text"},
			{Name: "serviceId", Type: "text"},
			{Name: "active", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "ord", Type: "integer", NotNull: true, Default: "0"},
			{Name: "source", Type: "text", NotNull: true, Default: "'manual'::text"},
			{Name: "approvedAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "discountPercent", Type: "integer"},
		},
		Constraints: []Constraint{
			{Name: "PromoItem_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "PromoItem_serviceId_fkey", Def: "FOREIGN KEY (\"serviceId\") REFERENCES \"Service\"(id) ON DELETE SET NULL"},
		},
	})
}
