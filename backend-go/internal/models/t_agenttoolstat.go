// موديل جدول "AgentToolStat" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AgentToolStat",
		Columns: []Column{
			{Name: "day", Type: "date", NotNull: true},
			{Name: "agent", Type: "text", NotNull: true},
			{Name: "tool", Type: "text", NotNull: true},
			{Name: "calls", Type: "integer", NotNull: true, Default: "0"},
			{Name: "failures", Type: "integer", NotNull: true, Default: "0"},
			{Name: "totalMs", Type: "bigint", NotNull: true, Default: "0"},
			{Name: "lastError", Type: "text"},
		},
		Constraints: []Constraint{
			{Name: "AgentToolStat_pkey", Def: "PRIMARY KEY (day, agent, tool)"},
		},
	})
}
