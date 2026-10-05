package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type AchievementVoice struct {
	AchievementID string `db:"achievementId"`
	FileKey       string `db:"fileKey"`
	ContentType   string `db:"contentType"`
	Seconds       int    `db:"seconds"`
}

type AchievementVoiceRepository struct{ db *sqlx.DB }

func NewAchievementVoiceRepository(db *sqlx.DB) *AchievementVoiceRepository {
	return &AchievementVoiceRepository{db: db}
}

func (r *AchievementVoiceRepository) Save(v AchievementVoice) error {
	_, err := r.db.Exec(`INSERT INTO "AchievementVoice" ("achievementId", "fileKey", "contentType", seconds)
		VALUES ($1, $2, $3, $4) ON CONFLICT ("achievementId") DO UPDATE
		SET "fileKey" = EXCLUDED."fileKey", "contentType" = EXCLUDED."contentType", seconds = EXCLUDED.seconds, "createdAt" = now()`,
		v.AchievementID, v.FileKey, v.ContentType, v.Seconds)
	return err
}

func (r *AchievementVoiceRepository) Get(id string) (*AchievementVoice, error) {
	var v AchievementVoice
	if err := r.db.Get(&v, `SELECT "achievementId", "fileKey", "contentType", seconds FROM "AchievementVoice" WHERE "achievementId" = $1`, id); err != nil {
		return nil, err
	}
	return &v, nil
}

// Seconds مدة الفويس لكل إنجاز عنده فويس.
func (r *AchievementVoiceRepository) Seconds(ids []string) map[string]int {
	out := map[string]int{}
	if len(ids) == 0 {
		return out
	}
	rows := []AchievementVoice{}
	_ = r.db.Select(&rows, `SELECT "achievementId", "fileKey", "contentType", seconds FROM "AchievementVoice" WHERE "achievementId" = ANY($1)`, pq.Array(ids))
	for _, v := range rows {
		out[v.AchievementID] = v.Seconds
	}
	return out
}
