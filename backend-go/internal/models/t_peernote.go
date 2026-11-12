// موديل جدول "PeerNote" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "PeerNote",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "checkinId", Type: "text", NotNull: true},
			{Name: "authorId", Type: "text", NotNull: true},
			{Name: "targetId", Type: "text", NotNull: true},
			{Name: "score", Type: "integer", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true, Default: "'NOTE'::text"},
			{Name: "text", Type: "text"},
			{Name: "needsFollowup", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "fWhat", Type: "text"},
			{Name: "fWhy", Type: "text"},
			{Name: "fWish", Type: "text"},
			{Name: "suggestion", Type: "text"},
			{Name: "accepted", Type: "boolean"},
			{Name: "acceptNote", Type: "text"},
			{Name: "severity", Type: "text", NotNull: true, Default: "'NORMAL'::text"},
			{Name: "reportedAt", Type: "timestamp with time zone"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "PeerNote_authorId_fkey", Def: "FOREIGN KEY (\"authorId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
			{Name: "PeerNote_check", Def: "CHECK ((\"authorId\" <> \"targetId\"))"},
			{Name: "PeerNote_checkinId_fkey", Def: "FOREIGN KEY (\"checkinId\") REFERENCES \"PeerCheckin\"(id) ON DELETE CASCADE"},
			{Name: "PeerNote_kind_check", Def: "CHECK ((kind = ANY (ARRAY['GOOD'::text, 'NOTE'::text, 'PROBLEM'::text])))"},
			{Name: "PeerNote_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "PeerNote_score_check", Def: "CHECK (((score >= 1) AND (score <= 5)))"},
			{Name: "PeerNote_severity_check", Def: "CHECK ((severity = ANY (ARRAY['NORMAL'::text, 'ATTENTION'::text, 'URGENT'::text])))"},
			{Name: "PeerNote_targetId_fkey", Def: "FOREIGN KEY (\"targetId\") REFERENCES \"Employee\"(id) ON DELETE CASCADE"},
		},
		Indexes: []string{
			"CREATE INDEX \"PeerNote_author_idx\" ON public.\"PeerNote\" USING btree (\"authorId\", \"createdAt\" DESC)",
			"CREATE INDEX \"PeerNote_target_idx\" ON public.\"PeerNote\" USING btree (\"targetId\", \"createdAt\" DESC)",
		},
	})
}
