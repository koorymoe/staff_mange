// موديل جدول "AiMetric" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiMetric",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "metricKey", Type: "text", NotNull: true},
			{Name: "scope", Type: "text", NotNull: true, Default: "'COMPANY'::text"},
			{Name: "scopeId", Type: "text"},
			{Name: "periodStart", Type: "date", NotNull: true},
			{Name: "periodEnd", Type: "date", NotNull: true},
			{Name: "value", Type: "numeric(14,3)", NotNull: true, Default: "0"},
			{Name: "sampleCount", Type: "integer", NotNull: true, Default: "0"},
			{Name: "details", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "computedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiMetric_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"AiMetric_unique\" ON public.\"AiMetric\" USING btree (\"metricKey\", scope, COALESCE(\"scopeId\", ''::text), \"periodStart\", \"periodEnd\")",
		},
	})
}
