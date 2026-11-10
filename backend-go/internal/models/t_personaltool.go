// موديل جدول "PersonalTool" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PersonalTool",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "barcode", Type: "text", NotNull: true},
			{Name: "status", Type: "\"ToolStatus\"", NotNull: true, Default: "'AVAILABLE'::\"ToolStatus\""},
			{Name: "checkedOut", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "PersonalTool_barcode_key", Def: "UNIQUE (barcode)"},
			{Name: "PersonalTool_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "PersonalTool_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"PersonalTool_employeeId_idx\" ON public.\"PersonalTool\" USING btree (\"employeeId\")",
		},
	})
}
