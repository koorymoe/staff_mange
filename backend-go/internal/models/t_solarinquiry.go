// موديل جدول "SolarInquiry" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SolarInquiry",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true, Default: "(gen_random_uuid())::text"},
			{Name: "number", Type: "integer", NotNull: true, Default: "nextval('\"SolarInquiry_number_seq\"'::regclass)"},
			{Name: "customerName", Type: "text", NotNull: true},
			{Name: "phone", Type: "text", NotNull: true},
			{Name: "area", Type: "text"},
			{Name: "note", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "techId", Type: "text"},
			{Name: "techClaimedAt", Type: "timestamp with time zone"},
			{Name: "techOutcome", Type: "text"},
			{Name: "techNote", Type: "text"},
			{Name: "techContactAt", Type: "timestamp with time zone"},
			{Name: "quoteGiven", Type: "boolean"},
			{Name: "quoteAmount", Type: "numeric"},
			{Name: "qualityById", Type: "text"},
			{Name: "qualityAt", Type: "timestamp with time zone"},
			{Name: "qCustomerReached", Type: "boolean"},
			{Name: "qTechCalled", Type: "text"},
			{Name: "qIntent", Type: "text"},
			{Name: "qGotQuote", Type: "text"},
			{Name: "qQuoteFair", Type: "text"},
			{Name: "qNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "SolarInquiry_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "SolarInquiry_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SolarInquiry_qGotQuote_check", Def: "CHECK ((\"qGotQuote\" = ANY (ARRAY['YES'::text, 'NO'::text, 'UNSURE'::text])))"},
			{Name: "SolarInquiry_qIntent_check", Def: "CHECK ((\"qIntent\" = ANY (ARRAY['JUST_ASKING'::text, 'WANTS_INSTALL'::text, 'NOT_INTERESTED'::text])))"},
			{Name: "SolarInquiry_qQuoteFair_check", Def: "CHECK ((\"qQuoteFair\" = ANY (ARRAY['YES'::text, 'NO'::text, 'UNSURE'::text])))"},
			{Name: "SolarInquiry_qTechCalled_check", Def: "CHECK ((\"qTechCalled\" = ANY (ARRAY['YES'::text, 'NO'::text, 'UNSURE'::text])))"},
			{Name: "SolarInquiry_qualityById_fkey", Def: "FOREIGN KEY (\"qualityById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "SolarInquiry_techId_fkey", Def: "FOREIGN KEY (\"techId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "SolarInquiry_techOutcome_check", Def: "CHECK ((\"techOutcome\" = ANY (ARRAY['JUST_ASKING'::text, 'WANTS_INSTALL'::text, 'NO_ANSWER'::text, 'NOT_INTERESTED'::text])))"},
		},
		Indexes: []string{
			"CREATE INDEX \"SolarInquiry_open_idx\" ON public.\"SolarInquiry\" USING btree (\"createdAt\" DESC) WHERE (\"qualityAt\" IS NULL)",
		},
	})
}
