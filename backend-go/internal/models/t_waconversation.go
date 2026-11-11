// موديل جدول "WaConversation" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "WaConversation",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "agent", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text"},
			{Name: "phone", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "followUpId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'BOT'::text"},
			{Name: "sandbox", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "humanById", Type: "text"},
			{Name: "lastMessageAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "lastReadAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "WaConversation_agent_check", Def: "CHECK ((agent = ANY (ARRAY['QUALITY'::text, 'BOOKING'::text])))"},
			{Name: "WaConversation_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "WaConversation_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE SET NULL"},
			{Name: "WaConversation_followUpId_fkey", Def: "FOREIGN KEY (\"followUpId\") REFERENCES \"QualityFollowUp\"(id) ON DELETE SET NULL"},
			{Name: "WaConversation_humanById_fkey", Def: "FOREIGN KEY (\"humanById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "WaConversation_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "WaConversation_status_check", Def: "CHECK ((status = ANY (ARRAY['BOT'::text, 'HUMAN'::text, 'CLOSED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"WaConversation_list_idx\" ON public.\"WaConversation\" USING btree (agent, status, \"lastMessageAt\" DESC)",
			"CREATE UNIQUE INDEX \"WaConversation_open_uq\" ON public.\"WaConversation\" USING btree (agent, phone) WHERE ((status <> 'CLOSED'::text) AND (NOT sandbox))",
		},
	})
}
