// موديل جدول "Project" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Project",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "code", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "rep", Type: "text"},
			{Name: "phone", Type: "text"},
			{Name: "location", Type: "text"},
			{Name: "workType", Type: "text"},
			{Name: "refPerson", Type: "text"},
			{Name: "stage", Type: "text", NotNull: true, Default: "'1. اتصال بالزبون'::text"},
			{Name: "price", Type: "text"},
			{Name: "staff", Type: "text"},
			{Name: "time", Type: "text"},
			{Name: "task", Type: "text"},
			{Name: "priority", Type: "text", NotNull: true, Default: "'عادي'::text"},
			{Name: "deliveryDate", Type: "text"},
			{Name: "survey", Type: "jsonb"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "bookingId", Type: "text"},
			{Name: "mapLatitude", Type: "double precision"},
			{Name: "mapLongitude", Type: "double precision"},
			{Name: "contractPdfBase64", Type: "text"},
			{Name: "signedContractPdfBase64", Type: "text"},
			{Name: "responsibleEmployeeId", Type: "text"},
			{Name: "surveyorEmployeeId", Type: "text"},
			{Name: "locationUrl", Type: "text"},
			{Name: "createdByEmployeeId", Type: "text"},
			{Name: "delegatedToEmployeeId", Type: "text"},
			{Name: "delegatedByEmployeeId", Type: "text"},
			{Name: "delegatedAt", Type: "timestamp with time zone"},
			{Name: "customerContactedAt", Type: "timestamp without time zone"},
			{Name: "customerContactedById", Type: "text"},
			{Name: "customerContactNote", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "Project_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE SET NULL"},
			{Name: "Project_code_key", Def: "UNIQUE (code)"},
			{Name: "Project_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "Project_responsibleEmployeeId_fkey", Def: "FOREIGN KEY (\"responsibleEmployeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "Project_surveyorEmployeeId_fkey", Def: "FOREIGN KEY (\"surveyorEmployeeId\") REFERENCES \"Employee\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Project_bookingId_idx\" ON public.\"Project\" USING btree (\"bookingId\")",
			"CREATE INDEX \"Project_createdAt_idx\" ON public.\"Project\" USING btree (\"createdAt\" DESC)",
			"CREATE INDEX \"Project_delegatedTo_idx\" ON public.\"Project\" USING btree (\"delegatedToEmployeeId\")",
			"CREATE INDEX \"Project_stage_idx\" ON public.\"Project\" USING btree (stage)",
		},
	})
}
