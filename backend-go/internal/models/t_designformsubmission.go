// موديل جدول "DesignFormSubmission" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "DesignFormSubmission",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "formId", Type: "text", NotNull: true},
			{Name: "answers", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "submittedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "DesignFormSubmission_formId_fkey", Def: "FOREIGN KEY (\"formId\") REFERENCES \"DesignForm\"(id) ON DELETE CASCADE"},
			{Name: "DesignFormSubmission_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"DesignFormSubmission_formId_idx\" ON public.\"DesignFormSubmission\" USING btree (\"formId\")",
		},
	})
}
