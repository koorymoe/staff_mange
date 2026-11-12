// موديل جدول "AiAgentName" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiAgentName",
		Columns: []Column{
			{Name: "agent", Type: "text", NotNull: true},
			{Name: "name", Type: "text", NotNull: true},
			{Name: "updatedById", Type: "text"},
			{Name: "updatedAt", Type: "timestamp with time zone", NotNull: true, Default: "now()"},
		},
		Constraints: []Constraint{
			{Name: "AiAgentName_agent_check", Def: "CHECK ((agent = ANY (ARRAY['QUALITY'::text, 'BOOKING'::text, 'MATRIX'::text, 'IRIS'::text])))"},
			{Name: "AiAgentName_pkey", Def: "PRIMARY KEY (agent)"},
			{Name: "AiAgentName_updatedById_fkey", Def: "FOREIGN KEY (\"updatedById\") REFERENCES \"Employee\"(id) ON DELETE SET NULL"},
		},
	})
}
