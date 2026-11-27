// موديل جدول "ScoreWeightConfig" — أوزان بنود التقييم لكل مجموعة (دازاي 11-27).

package models

func init() {
	register(Table{
		Name: "ScoreWeightConfig",
		Columns: []Column{
			{Name: "id", Type: "integer", NotNull: true, Default: "1"},
			{Name: "weights", Type: "jsonb", NotNull: true, Default: "'{}'::jsonb"},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "ScoreWeightConfig_pkey", Def: "PRIMARY KEY (id)"},
			{Name: "ScoreWeightConfig_one_row", Def: "CHECK ((id = 1))"},
			{Name: "ScoreWeightConfig_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
