// موديل جدول "ToolRequest" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ToolRequest",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text", NotNull: true},
			{Name: "toolId", Type: "text", NotNull: true},
			{Name: "status", Type: "\"ToolRequestStatus\"", NotNull: true, Default: "'PENDING'::\"ToolRequestStatus\""},
			{Name: "approvedById", Type: "text"},
			{Name: "requestedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "returnedAt", Type: "timestamp without time zone"},
			{Name: "approvedAt", Type: "timestamp without time zone"},
			{Name: "reason", Type: "text"},
			{Name: "description", Type: "text"},
			{Name: "purchasePrice", Type: "double precision"},
			{Name: "procurementRequestId", Type: "text"},
			{Name: "requestKind", Type: "text"},
			{Name: "rejectedAt", Type: "timestamp without time zone"},
			{Name: "rejectedById", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "ToolRequest_approvedById_fkey", Def: "FOREIGN KEY (\"approvedById\") REFERENCES \"Employee\"(id)"},
			{Name: "ToolRequest_employeeId_fkey", Def: "FOREIGN KEY (\"employeeId\") REFERENCES \"Employee\"(id)"},
			{Name: "ToolRequest_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ToolRequest_rejectedById_fkey", Def: "FOREIGN KEY (\"rejectedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ToolRequest_toolId_fkey", Def: "FOREIGN KEY (\"toolId\") REFERENCES \"OnDemandTool\"(id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ToolRequest_employeeId_idx\" ON public.\"ToolRequest\" USING btree (\"employeeId\")",
			"CREATE INDEX \"ToolRequest_toolId_idx\" ON public.\"ToolRequest\" USING btree (\"toolId\")",
		},
	})
}
