// موديل جدول "ProjectPayment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ProjectPayment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "projectId", Type: "text", NotNull: true},
			{Name: "amount", Type: "numeric(16,0)", NotNull: true},
			{Name: "paidAt", Type: "date", NotNull: true},
			{Name: "method", Type: "text", NotNull: true, Default: "'CASH'::text"},
			{Name: "receiptNo", Type: "text"},
			{Name: "note", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "verifiedById", Type: "text"},
			{Name: "verifiedAt", Type: "timestamp with time zone"},
			{Name: "cancelledAt", Type: "timestamp with time zone"},
			{Name: "cancelledById", Type: "text"},
			{Name: "cancelReason", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "ProjectPayment_amount_check", Def: "CHECK ((amount > (0)::numeric))"},
			{Name: "ProjectPayment_cancelledById_fkey", Def: "FOREIGN KEY (\"cancelledById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ProjectPayment_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ProjectPayment_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ProjectPayment_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE CASCADE"},
			{Name: "ProjectPayment_verifiedById_fkey", Def: "FOREIGN KEY (\"verifiedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
		Indexes: []string{
			"CREATE INDEX \"ProjectPayment_paid_idx\" ON public.\"ProjectPayment\" USING btree (\"paidAt\")",
			"CREATE INDEX \"ProjectPayment_project_idx\" ON public.\"ProjectPayment\" USING btree (\"projectId\")",
		},
	})
}
