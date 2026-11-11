// موديل جدول "PushSubscription" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PushSubscription",
		Columns: []Column{
			{Name: "endpoint", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "p256dh", Type: "text", NotNull: true},
			{Name: "auth", Type: "text", NotNull: true},
			{Name: "userAgent", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "PushSubscription_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "PushSubscription_pkey", Def: "PRIMARY KEY (endpoint)"},
		},
		Indexes: []string{
			"CREATE INDEX \"PushSubscription_employee_idx\" ON public.\"PushSubscription\" USING btree (\"employeeId\")",
		},
	})
}
