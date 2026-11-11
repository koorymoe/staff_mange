// موديل جدول "RoleMetricTarget" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "RoleMetricTarget",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "role", Type: "text", NotNull: true},
			{Name: "metricKey", Type: "text", NotNull: true},
			{Name: "targetValue", Type: "numeric(14,2)", NotNull: true},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedByName", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "RoleMetricTarget_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "RoleMetricTarget_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX \"RoleMetricTarget_role_metric_idx\" ON public.\"RoleMetricTarget\" USING btree (role, \"metricKey\")",
		},
	})
}
