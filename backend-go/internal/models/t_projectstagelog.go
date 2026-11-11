// موديل جدول "ProjectStageLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectStageLog",
		Columns: []Column{
			{Name: "id", Type: "bigint", NotNull: true, Default: "nextval('\"ProjectStageLog_id_seq\"'::regclass)"},
			{Name: "projectId", Type: "text", NotNull: true},
			{Name: "fromStage", Type: "text"},
			{Name: "toStage", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ProjectStageLog_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProjectStageLog_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"ProjectStageLog_project_idx\" ON public.\"ProjectStageLog\" USING btree (\"projectId\", \"createdAt\")",
		},
	})
}
