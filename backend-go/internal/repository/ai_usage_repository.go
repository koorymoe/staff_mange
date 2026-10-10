package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type AiUsageRepository struct{ db *sqlx.DB }

func NewAiUsageRepository(db *sqlx.DB) *AiUsageRepository { return &AiUsageRepository{db: db} }

func (r *AiUsageRepository) Add(feature, model string, in, out, cacheRead, cacheWrite int64) {
	_, _ = r.db.Exec(`INSERT INTO "AiUsage" (day, feature, model, calls, "inputTokens", "outputTokens", "cacheRead", "cacheWrite")
		VALUES (baghdad_today(), $1, $2, 1, $3, $4, $5, $6)
		ON CONFLICT (day, feature, model) DO UPDATE SET calls = "AiUsage".calls + 1,
		  "inputTokens" = "AiUsage"."inputTokens" + EXCLUDED."inputTokens",
		  "outputTokens" = "AiUsage"."outputTokens" + EXCLUDED."outputTokens",
		  "cacheRead" = "AiUsage"."cacheRead" + EXCLUDED."cacheRead",
		  "cacheWrite" = "AiUsage"."cacheWrite" + EXCLUDED."cacheWrite"`, feature, model, in, out, cacheRead, cacheWrite)
}

type AiUsageRow struct {
	Day          time.Time `db:"day" json:"day"`
	Feature      string    `db:"feature" json:"feature"`
	Model        string    `db:"model" json:"model"`
	Calls        int       `db:"calls" json:"calls"`
	InputTokens  int64     `db:"inputTokens" json:"inputTokens"`
	OutputTokens int64     `db:"outputTokens" json:"outputTokens"`
	CacheRead    int64     `db:"cacheRead" json:"cacheRead"`
	CacheWrite   int64     `db:"cacheWrite" json:"cacheWrite"`
}

func (r *AiUsageRepository) Since(days int) ([]AiUsageRow, error) {
	rows := []AiUsageRow{}
	err := r.db.Select(&rows, `SELECT * FROM "AiUsage" WHERE day > baghdad_today() - $1::int ORDER BY day DESC, feature`, days)
	return rows, err
}
