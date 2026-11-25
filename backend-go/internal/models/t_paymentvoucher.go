// موديل جدول "PaymentVoucher" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-25 دازاي: «خانة صرف السندات — لأي مشروع أو حجز أو مشروع خارجي (مو شرط حجز)، كلفة السند،
// ووقت الصرف، والمبلغ لمنو انسلّم لو انحط بالكاشير».
func init() {
	register(Table{
		Name: "PaymentVoucher",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "number", Type: "bigint", NotNull: true},
			{Name: "targetKind", Type: "text", NotNull: true},
			{Name: "projectId", Type: "text"},
			{Name: "bookingId", Type: "text"},
			{Name: "externalName", Type: "text"},
			{Name: "purpose", Type: "text", NotNull: true},
			{Name: "amount", Type: "numeric", NotNull: true},
			{Name: "paidAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "recipientKind", Type: "text", NotNull: true},
			{Name: "recipientEmployeeId", Type: "text"},
			{Name: "note", Type: "text"},
			{Name: "createdById", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "cancelledAt", Type: "timestamp with time zone"},
			{Name: "cancelledById", Type: "text"},
			{Name: "cancelReason", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "PaymentVoucher_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "PaymentVoucher_number_key", Def: "UNIQUE (number)"},
			{Name: "PaymentVoucher_projectId_fkey", Def: "FOREIGN KEY (\"projectId\") REFERENCES \"Project\"(id) ON DELETE SET NULL"},
			{Name: "PaymentVoucher_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "PaymentVoucher_recipientEmployeeId_fkey", Def: "FOREIGN KEY (\"recipientEmployeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "PaymentVoucher_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "PaymentVoucher_cancelledById_fkey", Def: "FOREIGN KEY (\"cancelledById\") REFERENCES \"Employee\"(id)"},
			{Name: "payment_voucher_amount_ok", Def: "CHECK ((amount > (0)::numeric))"},
			{Name: "payment_voucher_target_ok", Def: "CHECK ((((\"targetKind\" = 'PROJECT'::text) AND (\"projectId\" IS NOT NULL)) OR ((\"targetKind\" = 'BOOKING'::text) AND (\"bookingId\" IS NOT NULL)) OR ((\"targetKind\" = 'EXTERNAL'::text) AND (\"externalName\" IS NOT NULL))))"},
			{Name: "payment_voucher_recipient_ok", Def: "CHECK ((((\"recipientKind\" = 'EMPLOYEE'::text) AND (\"recipientEmployeeId\" IS NOT NULL)) OR (\"recipientKind\" = 'CASHIER'::text)))"},
		},
		Indexes: []string{"CREATE INDEX \"PaymentVoucher_paidAt_idx\" ON public.\"PaymentVoucher\" USING btree (\"paidAt\" DESC)"},
	})
}
