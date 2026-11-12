// موديل جدول "AiUsage" — عدّله هنا، والتحديث يتولّد بـ`go run ./cmd/makemigration`.

package models

func init() {
	register(Table{
		Name: "AiUsage",
		Columns: []Column{
			{Name: "day", Type: "date", NotNull: true},
			{Name: "feature", Type: "text", NotNull: true},
			{Name: "model", Type: "text", NotNull: true, Default: "''::text"},
			{Name: "calls", Type: "integer", NotNull: true, Default: "0"},
			{Name: "inputTokens", Type: "bigint", NotNull: true, Default: "0"},
			{Name: "outputTokens", Type: "bigint", NotNull: true, Default: "0"},
			{Name: "cacheRead", Type: "bigint", NotNull: true, Default: "0"},
			{Name: "cacheWrite", Type: "bigint", NotNull: true, Default: "0"},
		},
		Constraints: []Constraint{
			{Name: "AiUsage_pkey", Def: "PRIMARY KEY (day, feature, model)"},
		},
	})
}
