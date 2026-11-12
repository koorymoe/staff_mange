// موديل جدول "PrivacyPolicyAcceptanceLog" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PrivacyPolicyAcceptanceLog",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "acceptedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "pointsHash", Type: "text", NotNull: true},
			{Name: "pointsCount", Type: "integer", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "PrivacyPolicyAcceptanceLog_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "PrivacyPolicyAcceptanceLog_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"PrivacyPolicyAcceptanceLog_emp\" ON public.\"PrivacyPolicyAcceptanceLog\" USING btree (\"employeeId\", \"acceptedAt\" DESC)",
		},
	})
}
