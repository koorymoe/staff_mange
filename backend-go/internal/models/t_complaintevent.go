// موديل جدول "ComplaintEvent" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "ComplaintEvent",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "complaintId", Type: "text", NotNull: true},
			{Name: "kind", Type: "text", NotNull: true},
			{Name: "detail", Type: "text"},
			{Name: "byEmployeeId", Type: "text"},
			{Name: "byName", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ComplaintEvent_byEmployeeId_fkey", Def: "FOREIGN KEY (\"byEmployeeId\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
			{Name: "ComplaintEvent_complaintId_fkey", Def: "FOREIGN KEY (\"complaintId\") REFERENCES \"Complaint\"(id) ON DELETE CASCADE"},
			{Name: "ComplaintEvent_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"ComplaintEvent_complaint_time_idx\" ON public.\"ComplaintEvent\" USING btree (\"complaintId\", \"createdAt\" DESC)",
		},
	})
}
