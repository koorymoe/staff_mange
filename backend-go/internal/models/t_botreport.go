// موديل جدول "BotReport" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BotReport",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "number", Type: "integer", NotNull: true, Default: "nextval('\"BotReport_number_seq\"'::regclass)"},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "chatId", Type: "bigint", NotNull: true},
			{Name: "body", Type: "text", NotNull: true},
			{Name: "photos", Type: "text[]", NotNull: true, Default: "'{}'::text[]"},
			{Name: "mediaGroupId", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'NEW'::text"},
			{Name: "adminNote", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
			{Name: "source", Type: "text", NotNull: true, Default: "'telegram'::text"},
			{Name: "page", Type: "text"},
			{Name: "photosPurgedAt", Type: "timestamp without time zone"},
		},
		Constraints: []Constraint{
			{Name: "BotReport_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "BotReport_number_key", Def: "UNIQUE (number)"},
			{Name: "BotReport_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"BotReport_employee_idx\" ON public.\"BotReport\" USING btree (\"employeeId\", \"createdAt\" DESC)",
			"CREATE INDEX \"BotReport_status_idx\" ON public.\"BotReport\" USING btree (status, \"createdAt\" DESC)",
		},
	})
}
