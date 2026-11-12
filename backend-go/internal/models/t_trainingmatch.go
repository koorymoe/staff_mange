// موديل جدول "TrainingMatch" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "TrainingMatch",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "curriculumId", Type: "text"},
			{Name: "topic", Type: "text", NotNull: true},
			{Name: "reason", Type: "text"},
			{Name: "sourceBookingId", Type: "text"},
			{Name: "sourceRatingId", Type: "text"},
			{Name: "skillId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'OPEN'::text"},
			{Name: "dueAt", Type: "timestamp with time zone"},
			{Name: "overdueNotifiedAt", Type: "timestamp with time zone"},
			{Name: "dismissNote", Type: "text"},
			{Name: "startedAt", Type: "timestamp with time zone"},
			{Name: "doneAt", Type: "timestamp with time zone"},
			{Name: "reportText", Type: "text"},
			{Name: "reportScore", Type: "integer"},
			{Name: "reportedById", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "TrainingMatch_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "TrainingMatch_curriculumId_fkey", Def: "FOREIGN KEY (\"curriculumId\") REFERENCES \"TrainingCurriculum\"(id) ON DELETE SET NULL"},
			{Name: "TrainingMatch_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "TrainingMatch_kind_check", Def: "CHECK ((kind = ANY (ARRAY['BOOKING_ISSUE'::text, 'SKILL_GAP'::text, 'MANUAL'::text])))"},
			{Name: "TrainingMatch_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "TrainingMatch_reportScore_check", Def: "CHECK (((\"reportScore\" >= 1) AND (\"reportScore\" <= 5)))"},
			{Name: "TrainingMatch_reportedById_fkey", Def: "FOREIGN KEY (\"reportedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "TrainingMatch_status_check", Def: "CHECK ((status = ANY (ARRAY['OPEN'::text, 'IN_TRAINING'::text, 'DONE'::text, 'DISMISSED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"TrainingMatch_open_idx\" ON public.\"TrainingMatch\" USING btree (\"employeeId\") WHERE (status = ANY (ARRAY['OPEN'::text, 'IN_TRAINING'::text]))",
			"CREATE UNIQUE INDEX \"TrainingMatch_rating_uq\" ON public.\"TrainingMatch\" USING btree (\"sourceRatingId\") WHERE (\"sourceRatingId\" IS NOT NULL)",
			"CREATE UNIQUE INDEX \"TrainingMatch_skill_uq\" ON public.\"TrainingMatch\" USING btree (\"employeeId\", \"skillId\") WHERE (\"skillId\" IS NOT NULL)",
		},
	})
}
