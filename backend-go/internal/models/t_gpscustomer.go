// موديل جدول "GpsCustomer" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "GpsCustomer",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "fullName", Type: "text", NotNull: true},
			{Name: "fatherName", Type: "text"},
			{Name: "grandfatherName", Type: "text"},
			{Name: "phone", Type: "text", NotNull: true},
			{Name: "address", Type: "text"},
			{Name: "governorate", Type: "text"},
			{Name: "idCardFrontUrl", Type: "text"},
			{Name: "idCardBackUrl", Type: "text"},
			{Name: "residenceCardFrontUrl", Type: "text"},
			{Name: "residenceCardBackUrl", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "GpsCustomer_pkey", Def: "PRIMARY KEY (id)"},
		},
	})
}
