// موديل جدول "SolarIntakeConfig" — إعدادات استفسار الطاقة الشمسية (دازاي 11-27).

package models

func init() {
	register(Table{
		Name: "SolarIntakeConfig",
		Columns: []Column{
			{Name: "id", Type: "integer", NotNull: true, Default: "1"},
			{Name: "appliances", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "bands", Type: "jsonb", NotNull: true, Default: "'[]'::jsonb"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "SolarIntakeConfig_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "SolarIntakeConfig_one_row", Def: "CHECK ((id = 1))"},
			{Name: "SolarIntakeConfig_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
