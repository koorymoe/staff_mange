// موديل جدول "IntegrityRun" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "IntegrityRun",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "startedAt", Type: "timestamp with time zone", NotNull: true},
			{Name: "finishedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "results", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "problems", Type: "bigint", NotNull: true, Default: "0"},
			{Name: "critical", Type: "bigint", NotNull: true, Default: "0"},
			{Name: "failed", Type: "integer", NotNull: true, Default: "0"},
			{Name: "trigger", Type: "text", NotNull: true, Default: "'daily'::text"},
		},
		Constraints: []Constraint{
			{Name: "IntegrityRun_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"IntegrityRun_started\" ON public.\"IntegrityRun\" USING btree (\"startedAt\" DESC)",
		},
	})
}
