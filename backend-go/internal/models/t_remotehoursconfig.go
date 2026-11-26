// موديل جدول "RemoteHoursConfig" — فترة ساعات العمل من البيت وموعد رفعها (دازاي 11-26).

package models

func init() {
	register(Table{
		Name: "RemoteHoursConfig",
		Columns: []Column{
			{Name: "id", Type: "integer", NotNull: true, Default: "1"},
			{Name: "periodStartDay", Type: "integer", NotNull: true, Default: "26"},
			{Name: "closeDay", Type: "integer", NotNull: true, Default: "25"},
			{Name: "closeHour", Type: "integer", NotNull: true, Default: "23"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "RemoteHoursConfig_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "RemoteHoursConfig_one_row", Def: "CHECK ((id = 1))"},
			{Name: "RemoteHoursConfig_start_check", Def: "CHECK (((\"periodStartDay\" >= 1) AND (\"periodStartDay\" <= 28)))"},
			{Name: "RemoteHoursConfig_close_check", Def: "CHECK (((\"closeDay\" >= 1) AND (\"closeDay\" <= 28)))"},
			{Name: "RemoteHoursConfig_hour_check", Def: "CHECK (((\"closeHour\" >= 0) AND (\"closeHour\" <= 23)))"},
			{Name: "RemoteHoursConfig_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
