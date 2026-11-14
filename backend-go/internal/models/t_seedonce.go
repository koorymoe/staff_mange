// موديل جدول "SeedOnce" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (دازاي 11-15) «كل تشغيل يرجّع الي غيّرته»: البذور (أسعار، معايير، مهارات) تنزرع مرة وحدة
// بس — بعدها تعديل المدير أو حذفه يبقى. السطر هنا = هالبذرة انزرعت.
func init() {
	register(Table{
		Name: "SeedOnce",
		Columns: []Column{
			{Name: "key", Type: "text", NotNull: true},
			{Name: "doneAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SeedOnce_pkey", Def: "PRIMARY KEY (key)"},
		},
	})
}
