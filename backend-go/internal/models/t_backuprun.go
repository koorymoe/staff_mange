// موديل جدول "BackupRun" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BackupRun",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "startedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "finishedAt", Type: "timestamp with time zone"},
			{Name: "ok", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "fileName", Type: "text"},
			{Name: "sizeBytes", Type: "bigint", NotNull: true, Default: "0"},
			{Name: "tableCount", Type: "integer", NotNull: true, Default: "0"},
			{Name: "encrypted", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "offsite", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "offsiteTarget", Type: "text"},
			{Name: "hasUploads", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "hasEnv", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "warnings", Type: "text"},
			{Name: "error", Type: "text"},
			{Name: "keptCount", Type: "integer", NotNull: true, Default: "0"},
			{Name: "verified", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "verifyNote", Type: "text"},
			{Name: "verifyRows", Type: "bigint", NotNull: true, Default: "0"},
		},
		Constraints: []Constraint{
			{Name: "BackupRun_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"BackupRun_startedAt_idx\" ON public.\"BackupRun\" USING btree (\"startedAt\" DESC)",
		},
	})
}
