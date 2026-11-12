// موديل جدول "SimExercise" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimExercise",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "categoryId", Type: "text", NotNull: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "brief", Type: "text"},
			{Name: "engineKind", Type: "text", NotNull: true, Default: "'WIRING'::text"},
			{Name: "difficulty", Type: "integer", NotNull: true, Default: "1"},
			{Name: "timeLimitSec", Type: "integer"},
			{Name: "passScore", Type: "integer", NotNull: true, Default: "80"},
			{Name: "maxAttempts", Type: "integer"},
			{Name: "scene", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "steps", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "skillId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'DRAFT'::text"},
			{Name: "version", Type: "integer", NotNull: true, Default: "1"},
			{Name: "sourceRef", Type: "text"},
			{Name: "localPractice", Type: "text"},
			{Name: "verified", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "verifiedById", Type: "text"},
			{Name: "verifiedAt", Type: "timestamp with time zone"},
			{Name: "authorId", Type: "text"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
			{Name: "sortOrder", Type: "integer", NotNull: true, Default: "0"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimExercise_authorId_fkey", Def: "FOREIGN KEY (\"authorId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "SimExercise_categoryId_fkey", Def: "FOREIGN KEY (\"categoryId\") REFERENCES \"SimCategory\"(id) ON DELETE CASCADE"},
			{Name: "SimExercise_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SimExercise_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "SimExercise_skillId_fkey", Def: "FOREIGN KEY (\"skillId\") REFERENCES \"Skill\"(id) ON DELETE SET NULL"},
			{Name: "SimExercise_verifiedById_fkey", Def: "FOREIGN KEY (\"verifiedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"SimExercise_category_idx\" ON public.\"SimExercise\" USING btree (\"categoryId\", status, verified, \"sortOrder\")",
		},
	})
}
