// موديل جدول "Mission" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "Mission",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "code", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "stage", Type: "\"MissionStage\"", NotNull: true, Default: "'ASSIGNED'::\"MissionStage\""},
			{Name: "leaderId", Type: "text", NotNull: true},
			{Name: "memberIds", Type: "text[]", Default: "'{}'::text[]"},
			{Name: "customerLat", Type: "double precision"},
			{Name: "customerLng", Type: "double precision"},
			{Name: "customerAddress", Type: "text"},
			{Name: "assignedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "materialsReadyAt", Type: "timestamp without time zone"},
			{Name: "departedAt", Type: "timestamp without time zone"},
			{Name: "arrivedAt", Type: "timestamp without time zone"},
			{Name: "workStartedAt", Type: "timestamp without time zone"},
			{Name: "completedAt", Type: "timestamp without time zone"},
			{Name: "stoppedAt", Type: "timestamp without time zone"},
			{Name: "departureLat", Type: "double precision"},
			{Name: "departureLng", Type: "double precision"},
			{Name: "arrivalLat", Type: "double precision"},
			{Name: "arrivalLng", Type: "double precision"},
			{Name: "estimatedMinutes", Type: "integer"},
			{Name: "actualMinutes", Type: "integer"},
			{Name: "distanceKm", Type: "double precision"},
			{Name: "stopReason", Type: "text"},
			{Name: "notes", Type: "text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "updatedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "Mission_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id)"},
			{Name: "Mission_code_key", Def: "UNIQUE (code)"},
			{Name: "Mission_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"Mission_bookingId_idx\" ON public.\"Mission\" USING btree (\"bookingId\")",
			"CREATE INDEX \"Mission_leaderId_idx\" ON public.\"Mission\" USING btree (\"leaderId\")",
			"CREATE INDEX \"Mission_stage_idx\" ON public.\"Mission\" USING btree (stage)",
		},
	})
}
