// موديل جدول "BookingStageProof" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "BookingStageProof",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "bookingId", Type: "text", NotNull: true},
			{Name: "stage", Type: "text", NotNull: true},
			{Name: "employeeId", Type: "text"},
			{Name: "pressedAt", Type: "timestamp without time zone", NotNull: true},
			{Name: "receivedAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
			{Name: "offline", Type: "boolean", NotNull: true, Default: "false"},
			{Name: "lat", Type: "double precision"},
			{Name: "lng", Type: "double precision"},
			{Name: "accuracy", Type: "double precision"},
			{Name: "distanceM", Type: "double precision"},
			{Name: "verdict", Type: "text", NotNull: true, Default: "'OK'::text"},
		},
		Constraints: []Constraint{
			{Name: "BookingStageProof_bookingId_fkey", Def: "FOREIGN KEY (\"bookingId\") REFERENCES \"Booking\"(id) ON DELETE CASCADE"},
			{Name: "BookingStageProof_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX stage_proof_booking ON public.\"BookingStageProof\" USING btree (\"bookingId\")",
			"CREATE INDEX stage_proof_emp ON public.\"BookingStageProof\" USING btree (\"employeeId\", \"pressedAt\")",
		},
	})
}
