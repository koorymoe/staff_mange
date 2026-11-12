// موديل جدول "StoryInstance" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "StoryInstance",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "eventId", Type: "text", NotNull: true},
			{Name: "eventKind", Type: "text", NotNull: true},
			{Name: "storyType", Type: "text", NotNull: true},
			{Name: "version", Type: "integer", NotNull: true, Default: "1"},
			{Name: "senderEmployeeId", Type: "text"},
			{Name: "senderName", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "recipientEmployeeId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'QUEUED'::text"},
			{Name: "priority", Type: "integer", NotNull: true, Default: "50"},
			{Name: "physical", Type: "boolean", NotNull: true, Default: "true"},
			{Name: "groupKey", Type: "text"},
			{Name: "currentStep", Type: "integer", NotNull: true, Default: "0"},
			{Name: "payload", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "deliveredAt", Type: "timestamp with time zone"},
			{Name: "seenAt", Type: "timestamp with time zone"},
			{Name: "openedAt", Type: "timestamp with time zone"},
			{Name: "acknowledgedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "expiresAt", Type: "timestamp with time zone"},
			{Name: "recipientRef", Type: "text", NotNull: true},
			{Name: "recipientName", Type: "text", NotNull: true, Default: "''::text"},
		},
		Constraints: []Constraint{
			{Name: "StoryInstance_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "StoryInstance_recipientEmployeeId_fkey", Def: "FOREIGN KEY (\"recipientEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "StoryInstance_senderEmployeeId_fkey", Def: "FOREIGN KEY (\"senderEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "story_status_valid", Def: "CHECK (((status = ANY (ARRAY['QUEUED'::text, 'DELIVERED'::text, 'PLAYING'::text, 'SEEN'::text, 'OPENED'::text, 'ACKNOWLEDGED'::text, 'FAILED'::text])) AND ((status <> 'ACKNOWLEDGED'::text) OR (\"acknowledgedAt\" IS NOT NULL))))"},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX story_event_recipient_unique_v2 ON public.\"StoryInstance\" USING btree (\"eventKind\", \"eventId\", \"recipientRef\")",
			"CREATE UNIQUE INDEX story_group_open_ref_unique ON public.\"StoryInstance\" USING btree (\"recipientRef\", \"groupKey\") WHERE ((\"groupKey\" IS NOT NULL) AND (status = ANY (ARRAY['QUEUED'::text, 'DELIVERED'::text, 'PLAYING'::text])))",
			"CREATE INDEX story_pending_ref_idx ON public.\"StoryInstance\" USING btree (\"recipientRef\", priority DESC, \"createdAt\") WHERE (status = ANY (ARRAY['QUEUED'::text, 'DELIVERED'::text, 'PLAYING'::text]))",
		},
	})
}
