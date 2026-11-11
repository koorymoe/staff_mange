// موديل جدول "AdminEditLog" — 🏴‍☠️ ون بيس 11-10: كل تعديل/إضافة من لوحة الأدمن (منو، متى، قبل وبعد).
// أول جدول انكتب بالموديل أولاً، وتحديثه انولد بـcmd/makemigration.

package models

func init() {
	register(Table{
		Name: "AdminEditLog",
		Columns: []Column{
			{Name: "id", Type: "bigint", NotNull: true, Default: "nextval('\"AdminEditLog_id_seq\"'::regclass)"},
			{Name: "tableName", Type: "text", NotNull: true},
			{Name: "action", Type: "text", NotNull: true},
			{Name: "rowKey", Type: "jsonb", NotNull: true},
			{Name: "before", Type: "jsonb"},
			{Name: "after", Type: "jsonb"},
			{Name: "byId", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AdminEditLog_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"AdminEditLog_table_idx\" ON public.\"AdminEditLog\" USING btree (\"tableName\", \"createdAt\" DESC)",
		},
	})
}
