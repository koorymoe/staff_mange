// موديل جدول "DeleteLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (دازاي 11-15) «لازم كلشي يتسجل»: كل حذف نهائي لشي مهم (عرض سعر، مورد، منتج، خدمة، سيارة،
// مشروع…) ينحفظ قبلها نسخة كاملة من السطر: منو حذف، ومتى، ومن أي شاشة — وتكدر ترجعله.
func init() {
	register(Table{
		Name: "DeleteLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "tableName", Type: "text", NotNull: true},
			{Name: "rowId", Type: "text", NotNull: true},
			{Name: "actorId", Type: "text"},
			{Name: "route", Type: "text", NotNull: true},
			{Name: "data", Type: "jsonb", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "DeleteLog_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"DeleteLog_table_row_idx\" ON public.\"DeleteLog\" USING btree (\"tableName\", \"rowId\")",
		},
	})
}
