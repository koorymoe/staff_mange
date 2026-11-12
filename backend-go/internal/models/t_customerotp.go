// موديل جدول "CustomerOtp" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "CustomerOtp",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "accountId", Type: "text", NotNull: true},
			{Name: "purpose", Type: "text", NotNull: true},
			{Name: "codeHash", Type: "text", NotNull: true},
			{Name: "attempts", Type: "integer", NotNull: true, Default: "0"},
			{Name: "expiresAt", Type: "timestamp without time zone", NotNull: true},
			{Name: "usedAt", Type: "timestamp without time zone"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "CustomerOtp_accountId_fkey", Def: "FOREIGN KEY (\"accountId\") REFERENCES \"CustomerAccount\"(id) ON DELETE CASCADE"},
			{Name: "CustomerOtp_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX customer_otp_acc ON public.\"CustomerOtp\" USING btree (\"accountId\", \"createdAt\" DESC)",
		},
	})
}
