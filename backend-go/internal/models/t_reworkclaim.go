// موديل جدول "ReworkClaim" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

// (ل) 11-24 دازاي: «كادر الصبح طلع غلط وكادر المسا رجع صلّحه — وقت التصحيح ينخصم من الصبح وينضاف للمسا».
// الليدر الي صلّح يسجّل، المراقب يوافق، والساعات من الحجز (الوصول/البداية ← الإنهاء) × سعر ساعة الي صلّح.
func init() {
	register(Table{
		Name: "ReworkClaim",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "fixBookingId", Type: "text"},
			{Name: "originalBookingId", Type: "text", NotNull: true},
			{Name: "note", Type: "text", NotNull: true},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "hours", Type: "numeric"},
			{Name: "amount", Type: "numeric"},
			{Name: "createdById", Type: "text", NotNull: true},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "decidedById", Type: "text"},
			{Name: "decidedAt", Type: "timestamp with time zone"},
			{Name: "decisionNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "ReworkClaim_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ReworkClaim_fixBookingId_key", Def: "UNIQUE (\"fixBookingId\")"},
			{Name: "ReworkClaim_fixBookingId_fkey", Def: "FOREIGN KEY (\"fixBookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "ReworkClaim_originalBookingId_fkey", Def: "FOREIGN KEY (\"originalBookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "ReworkClaim_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "ReworkClaim_decidedById_fkey", Def: "FOREIGN KEY (\"decidedById\") REFERENCES \"Employee\"(id)"},
			{Name: "rework_claim_status_ok", Def: "CHECK ((status = ANY (ARRAY['PENDING'::text, 'APPROVED'::text, 'REJECTED'::text])))"},
			{Name: "rework_claim_distinct", Def: "CHECK ((\"fixBookingId\" <> \"originalBookingId\"))"},
		},
	})
}
