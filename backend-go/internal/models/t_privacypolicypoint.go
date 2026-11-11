// موديل جدول "PrivacyPolicyPoint" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PrivacyPolicyPoint",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "content", Type: "text", NotNull: true},
			{Name: "order", Type: "integer", NotNull: true, Default: "0"},
			{Name: "isActive", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "createdByEmployeeId", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "principleNo", Type: "integer"},
			{Name: "articleNo", Type: "integer"},
			{Name: "clauseTitle", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "PrivacyPolicyPoint_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"PrivacyPolicyPoint_order_idx\" ON public.\"PrivacyPolicyPoint\" USING btree (\"order\", \"createdAt\")",
			"CREATE UNIQUE INDEX \"PrivacyPolicyPoint_principle_uq\" ON public.\"PrivacyPolicyPoint\" USING btree (\"principleNo\") WHERE (\"principleNo\" IS NOT NULL)",
		},
	})
}
