// موديل جدول "JobDurationSample" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "JobDurationSample",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "systemName", Type: "text", NotNull: true},
			{Name: "jobType", Type: "text", NotNull: true},
			{Name: "itemCount", Type: "integer", NotNull: true},
			{Name: "crewSize", Type: "integer", NotNull: true},
			{Name: "durationMinutes", Type: "integer", NotNull: true},
			{Name: "bookingId", Type: "text"},
			{Name: "deviceMaintenanceTicketId", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "JobDurationSample_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id)"},
			{Name: "JobDurationSample_deviceMaintenanceTicketId_fkey", Def: "FOREIGN KEY (\"deviceMaintenanceTicketId\") REFERENCES \"DeviceMaintenanceTicket\"(id)"},
			{Name: "JobDurationSample_jobType_check", Def: "CHECK ((\"jobType\" = ANY (ARRAY['INSTALL'::text, 'MAINTENANCE'::text])))"},
			{Name: "JobDurationSample_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"JobDurationSample_system_jobtype_idx\" ON public.\"JobDurationSample\" USING btree (\"systemName\", \"jobType\")",
		},
	})
}
