// موديل جدول "CustomerAccount" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CustomerAccount",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "username", Type: "text", NotNull: true},
			{Name: "email", Type: "text", NotNull: true},
			{Name: "passwordHash", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "phone", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text"},
			{Name: "emailVerifiedAt", Type: "timestamp without time zone"},
			{Name: "sessionsInvalidatedAt", Type: "timestamp without time zone"},
			{Name: "lastSeenPromoAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "city", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "CustomerAccount_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE SET NULL"},
			{Name: "CustomerAccount_email_key", Def: "UNIQUE (email)"},
			{Name: "CustomerAccount_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "CustomerAccount_username_key", Def: "UNIQUE (username)"},
		},
	})
}
