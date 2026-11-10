// موديل جدول "MatrixCourier" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "MatrixCourier",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "senderId", Type: "text"},
			{Name: "recipientId", Type: "text", NotNull: true},
			{Name: "message", Type: "text", NotNull: true},
			{Name: "link", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "deliveredAt", Type: "timestamp without time zone"},
			{Name: "seenAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "MatrixCourier_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "MatrixCourier_recipientId_fkey", Def: "FOREIGN KEY (\"recipientId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "MatrixCourier_senderId_fkey", Def: "FOREIGN KEY (\"senderId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX matrix_courier_recipient ON public.\"MatrixCourier\" USING btree (\"recipientId\", \"createdAt\" DESC)",
			"CREATE INDEX matrix_courier_sender ON public.\"MatrixCourier\" USING btree (\"senderId\", \"createdAt\" DESC)",
		},
	})
}
