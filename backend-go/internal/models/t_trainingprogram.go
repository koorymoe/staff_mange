// موديل جدول "TrainingProgram" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingProgram",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "level", Type: "text", NotNull: true, Default: "'مبتدئ'::text"},
			{Name: "durationDays", Type: "integer", NotNull: true, Default: "1"},
			{Name: "startDate", Type: "date"},
			{Name: "endDate", Type: "date"},
			{Name: "targetDepartment", Type: "text"},
			{Name: "instructorId", Type: "text"},
			{Name: "objectives", Type: "text"},
			{Name: "content", Type: "text"},
			{Name: "passRate", Type: "integer", NotNull: true, Default: "80"},
			{Name: "cost", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "status", Type: "text", NotNull: true, Default: "'قيد التخطيط'::text"},
			{Name: "progress", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "TrainingProgram_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "TrainingProgram_instructorId_fkey", Def: "FOREIGN KEY (\"instructorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "TrainingProgram_level_check", Def: "CHECK ((level = ANY (ARRAY['مبتدئ'::text, 'متوسط'::text, 'متقدم'::text])))"},
			{Name: "TrainingProgram_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "TrainingProgram_progress_check", Def: "CHECK (((progress >= 0) AND (progress <= 100)))"},
			{Name: "TrainingProgram_status_check", Def: "CHECK ((status = ANY (ARRAY['قيد التخطيط'::text, 'جاري التنفيذ'::text, 'مكتمل'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"TrainingProgram_status_idx\" ON public.\"TrainingProgram\" USING btree (status)",
		},
	})
}
