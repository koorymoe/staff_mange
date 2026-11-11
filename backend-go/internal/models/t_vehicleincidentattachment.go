// موديل جدول "VehicleIncidentAttachment" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "VehicleIncidentAttachment",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "incidentId", Type: "text", NotNull: true},
			{Name: "url", Type: "text", NotNull: true},
			{Name: "mediaType", Type: "text", NotNull: true, Default: "'IMAGE'::text"},
			{Name: "createdAt", Type: "timestamp without time zone", NotNull: true, Default: "CURRENT_TIMESTAMP"},
		},
		Constraints: []Constraint{
			{Name: "VehicleIncidentAttachment_incidentId_fkey", Def: "FOREIGN KEY (\"incidentId\") REFERENCES \"VehicleIncident\"(id) ON DELETE CASCADE"},
			{Name: "VehicleIncidentAttachment_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"VehicleIncidentAttachment_incidentId_idx\" ON public.\"VehicleIncidentAttachment\" USING btree (\"incidentId\")",
		},
	})
}
