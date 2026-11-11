// موديل جدول "EmployeeProfileSurvey" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "EmployeeProfileSurvey",
		Columns: []Column{
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "major", Type: "text", NotNull: true},
			{Name: "degree", Type: "text", NotNull: true},
			{Name: "selfSalaryIQD", Type: "numeric"},
			{Name: "monthlyHours", Type: "integer"},
			{Name: "currentJob", Type: "text", NotNull: true},
			{Name: "companyRating", Type: "integer", NotNull: true},
			{Name: "companyNote", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "contractEnd", Type: "date"},
			{Name: "mismatch", Type: "text"},
			{Name: "submittedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "reviewedById", Type: "text"},
			{Name: "reviewedAt", Type: "timestamp with time zone"},
		},
		Constraints: []Constraint{
			{Name: "EmployeeProfileSurvey_companyRating_check", Def: "CHECK (((\"companyRating\" >= 1) AND (\"companyRating\" <= 5)))"},
			{Name: "EmployeeProfileSurvey_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "EmployeeProfileSurvey_monthlyHours_check", Def: "CHECK (((\"monthlyHours\" IS NULL) OR ((\"monthlyHours\" >= 1) AND (\"monthlyHours\" <= 400))))"},
			{Name: "EmployeeProfileSurvey_pkey", Def: "PRIMARY KEY (\"employeeId\")"},
			{Name: "EmployeeProfileSurvey_reviewedById_fkey", Def: "FOREIGN KEY (\"reviewedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "EmployeeProfileSurvey_selfSalaryIQD_check", Def: "CHECK (((\"selfSalaryIQD\" IS NULL) OR (\"selfSalaryIQD\" >= (0)::numeric)))"},
		},
	})
}
