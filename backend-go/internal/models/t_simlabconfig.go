// موديل جدول "SimLabConfig" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SimLabConfig",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "'singleton'::text"},
			{Name: "published", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "audienceRoles", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "safetyNotice", Type: "text", NotNull: true, Default: "'هذا تدريب محاكاة. لا تعتمد عليه بالميدان — الرجوع لكتالوگ الشركة المصنّعة إلزامي قبل أي توصيل حقيقي.'::text"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SimLabConfig_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SimLabConfig_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
