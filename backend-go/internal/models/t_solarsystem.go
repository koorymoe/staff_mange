// موديل جدول "SolarSystem" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "SolarSystem",
		Columns: []Column{
			{Name: "id", Type: "text", NotNull: true},
			{Name: "brand", Type: "text", NotNull: true},
			{Name: "model", Type: "text", NotNull: true},
			{Name: "capacity", Type: "text", NotNull: true},
			{Name: "panelId", Type: "text"},
			{Name: "panelQty", Type: "integer", NotNull: true, Default: "0"},
			{Name: "inverterId", Type: "text"},
			{Name: "inverterQty", Type: "integer", NotNull: true, Default: "0"},
			{Name: "batteryId", Type: "text"},
			{Name: "batteryQty", Type: "integer", NotNull: true, Default: "0"},
			{Name: "boardId", Type: "text"},
			{Name: "wiringDetails", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "wiringTotalCost", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "ironDetails", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "ironTotalCost", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "installPrice", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "programPrice", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "warrantyPrice", Type: "double precision", NotNull: true, Default: "0"},
			{Name: "notes", Type: "text"},
			{Name: "createdById", Type: "text"},
			{Name: "createdAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SolarSystem_batteryId_fkey", Def: "FOREIGN KEY (\"batteryId\") REFERENCES \"SolarComponent\"(id) ON DELETE RESTRICT"},
			{Name: "SolarSystem_boardId_fkey", Def: "FOREIGN KEY (\"boardId\") REFERENCES \"SolarComponent\"(id) ON DELETE RESTRICT"},
			{Name: "SolarSystem_createdById_fkey", Def: "FOREIGN KEY (\"createdById\") REFERENCES \"Employee\"(id)"},
			{Name: "SolarSystem_inverterId_fkey", Def: "FOREIGN KEY (\"inverterId\") REFERENCES \"SolarComponent\"(id) ON DELETE RESTRICT"},
			{Name: "SolarSystem_panelId_fkey", Def: "FOREIGN KEY (\"panelId\") REFERENCES \"SolarComponent\"(id) ON DELETE RESTRICT"},
			{Name: "SolarSystem_pkey", Def: "PRIMARY KEY (id)"},
		},
		Indexes: []string{
			"CREATE INDEX \"SolarSystem_battery_idx\" ON public.\"SolarSystem\" USING btree (\"batteryId\")",
			"CREATE INDEX \"SolarSystem_board_idx\" ON public.\"SolarSystem\" USING btree (\"boardId\")",
			"CREATE INDEX \"SolarSystem_brand_idx\" ON public.\"SolarSystem\" USING btree (brand)",
			"CREATE INDEX \"SolarSystem_inverter_idx\" ON public.\"SolarSystem\" USING btree (\"inverterId\")",
			"CREATE INDEX \"SolarSystem_panel_idx\" ON public.\"SolarSystem\" USING btree (\"panelId\")",
		},
	})
}
