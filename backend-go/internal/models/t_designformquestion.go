// موديل جدول "DesignFormQuestion" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DesignFormQuestion",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "label", Type: "text", NotNull: true},
			{Name: "type", Type: "text", NotNull: true},
			{Name: "options", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "required", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "order", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "formId", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "DesignFormQuestion_formId_fkey", Def: "FOREIGN KEY (\"formId\") REFERENCES \"DesignForm\"(id) ON DELETE CASCADE"},
			{Name: "DesignFormQuestion_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"DesignFormQuestion_formId_idx\" ON public.\"DesignFormQuestion\" USING btree (\"formId\")",
			"CREATE INDEX \"DesignFormQuestion_order_idx\" ON public.\"DesignFormQuestion\" USING btree (\"order\")",
		},
	})
}
