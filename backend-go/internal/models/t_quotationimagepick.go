// موديل جدول "QuotationImagePick" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "QuotationImagePick",
		Columns: []Column{
			{Name: "quotationId", Type: "text", NotNull: true},
			{Name: "chapter", Type: "text", NotNull: true},
			{Name: "imageId", Type: "text", NotNull: true},
			{Name: "pickedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "QuotationImagePick_imageId_fkey", Def: "FOREIGN KEY (\"imageId\") REFERENCES \"WorkTypeImage\"(id) ON DELETE CASCADE"},
			{Name: "QuotationImagePick_pkey", Def: "PRIMARY KEY (\"quotationId\", chapter)"},
			{Name: "QuotationImagePick_quotationId_fkey", Def: "FOREIGN KEY (\"quotationId\") REFERENCES \"Quotation\"(id) ON DELETE CASCADE"},
		},
	})
}
