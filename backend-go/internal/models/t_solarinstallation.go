// موديل جدول "SolarInstallation" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SolarInstallation",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "systemId", Type: "text", NotNull: true},
			{Name: "customerId", Type: "text", NotNull: true},
			{Name: "installDate", Type: "date", NotNull: true},
			{Name: "followUpAt", Type: "date", NotNull: true},
			{Name: "contactedAt", Type: "timestamp with time zone"},
			{Name: "contactedById", Type: "text"},
			{Name: "contactNotes", Type: "text"},
			{Name: "status", Type: "text", NotNull: true, Default: "'PENDING'::text"},
			{Name: "totalPrice", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "priceBreakdown", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "notes", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SolarInstallation_contactedById_fkey", Def: "FOREIGN KEY (\"contactedById\") REFERENCES \"Employee\"(id)"},
			{Name: "SolarInstallation_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "SolarInstallation_customerId_fkey", Def: "FOREIGN KEY (\"customerId\") REFERENCES \"Customer\"(id) ON DELETE RESTRICT"},
			{Name: "SolarInstallation_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SolarInstallation_status_check", Def: "CHECK ((status = ANY (ARRAY['PENDING'::text, 'CONTACTED'::text])))"},
			{Name: "SolarInstallation_systemId_fkey", Def: "FOREIGN KEY (\"systemId\") REFERENCES \"SolarSystem\"(id) ON DELETE RESTRICT"},
		},
		Indexes: []string{
			"CREATE INDEX \"SolarInstallation_customer_idx\" ON public.\"SolarInstallation\" USING btree (\"customerId\")",
			"CREATE INDEX \"SolarInstallation_due_idx\" ON public.\"SolarInstallation\" USING btree (\"followUpAt\") WHERE (status = 'PENDING'::text)",
			"CREATE INDEX \"SolarInstallation_system_idx\" ON public.\"SolarInstallation\" USING btree (\"systemId\")",
		},
	})
}
