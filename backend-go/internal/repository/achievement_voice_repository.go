package repository

import (
	"encoding/json"
	"time"

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

// ═══ تحليل الفويس (النص من Whisper المحلي) ═══

type VoiceTask struct {
	What  string `json:"what"`
	Count int    `json:"count"`
}

type VoiceAnalysis struct {
	AchievementID string      `db:"achievementId" json:"achievementId"`
	Status        string      `db:"analysisStatus" json:"status"`
	Transcript    *string     `db:"transcript" json:"transcript"`
	Summary       *string     `db:"summary" json:"summary"`
	Tasks         []byte      `db:"tasks" json:"-"`
	TaskList      []VoiceTask `db:"-" json:"tasks"`
	Style         *string     `db:"style" json:"style"`
	Error         *string     `db:"analysisError" json:"error"`
	AnalyzedAt    *time.Time  `db:"analyzedAt" json:"analyzedAt"`
}

func (r *AchievementVoiceRepository) Analysis(id string) (*VoiceAnalysis, error) {
	var v VoiceAnalysis
	if err := r.db.Get(&v, `SELECT "achievementId", "analysisStatus", transcript, summary, tasks, style, "analysisError", "analyzedAt"
		FROM "AchievementVoice" WHERE "achievementId" = $1`, id); err != nil {
		return nil, err
	}
	v.TaskList = []VoiceTask{}
	if len(v.Tasks) > 0 {
		_ = json.Unmarshal(v.Tasks, &v.TaskList)
	}
	return &v, nil
}

// Pending فويسات تنتظر تحليل (أقل من ٣ محاولات) — الأقدم أول.
func (r *AchievementVoiceRepository) Pending(limit int) ([]AchievementVoice, error) {
	rows := []AchievementVoice{}
	err := r.db.Select(&rows, `SELECT "achievementId", "fileKey", "contentType", seconds FROM "AchievementVoice"
		WHERE "analysisStatus" = 'PENDING' AND attempts < 3 ORDER BY "createdAt" LIMIT $1`, limit)
	return rows, err
}

func (r *AchievementVoiceRepository) MarkAttempt(id, errMsg string) {
	_, _ = r.db.Exec(`UPDATE "AchievementVoice" SET attempts = attempts + 1, "analysisError" = $2,
		"analysisStatus" = CASE WHEN attempts + 1 >= 3 THEN 'FAILED' ELSE 'PENDING' END WHERE "achievementId" = $1`, id, errMsg)
}

func (r *AchievementVoiceRepository) SaveAnalysis(id, transcript, summary, style string, tasks []VoiceTask) error {
	if tasks == nil {
		tasks = []VoiceTask{}
	}
	tj, _ := json.Marshal(tasks)
	_, err := r.db.Exec(`UPDATE "AchievementVoice" SET "analysisStatus" = 'DONE', transcript = $2, summary = NULLIF($3, ''),
		style = NULLIF($4, ''), tasks = $5, "analysisError" = NULL, "analyzedAt" = now() WHERE "achievementId" = $1`,
		id, transcript, summary, style, tj)
	return err
}

// ResetAnalysis فويس جديد (إعادة تسجيل) يرجع ينتظر التحليل.
func (r *AchievementVoiceRepository) ResetAnalysis(id string) {
	_, _ = r.db.Exec(`UPDATE "AchievementVoice" SET "analysisStatus" = 'PENDING', attempts = 0, transcript = NULL, summary = NULL,
		tasks = NULL, style = NULL, "analysisError" = NULL, "analyzedAt" = NULL WHERE "achievementId" = $1`, id)
}

func (r *AchievementVoiceRepository) AnalyzedCount() (done, pending int) {
	_ = r.db.Get(&done, `SELECT count(*) FROM "AchievementVoice" WHERE "analysisStatus" = 'DONE'`)
	_ = r.db.Get(&pending, `SELECT count(*) FROM "AchievementVoice" WHERE "analysisStatus" = 'PENDING'`)
	return
}

// AchievementEmployee صاحب الإنجاز — لشيل اسمه وأسماء زملاؤه من النص.
func (r *AchievementVoiceRepository) Names() []string {
	names := []string{}
	_ = r.db.Select(&names, `SELECT name FROM "Employee" WHERE length(btrim(name)) >= 3`)
	return names
}

// CustomerNameFor اسم زبون الحجز المربوط بالإنجاز (إذا أكو).
func (r *AchievementVoiceRepository) CustomerNameFor(achievementID string) string {
	var n string
	_ = r.db.Get(&n, `SELECT COALESCE(c.name, '') FROM "Achievement" a JOIN "Booking" b ON b.id = a."bookingId"
		JOIN "Customer" c ON c.id = b."customerId" WHERE a.id = $1`, achievementID)
	return n
}

// ═══ قاموس اللهجة ═══

type IraqiPhrase struct {
	Phrase    string    `db:"phrase" json:"phrase"`
	Meaning   string    `db:"meaning" json:"meaning"`
	Uses      int       `db:"uses" json:"uses"`
	FirstSeen time.Time `db:"firstSeen" json:"firstSeen"`
	LastSeen  time.Time `db:"lastSeen" json:"lastSeen"`
}

func (r *AchievementVoiceRepository) LearnPhrase(phrase, meaning string) {
	_, _ = r.db.Exec(`INSERT INTO "IraqiPhrase" (phrase, meaning) VALUES ($1, $2)
		ON CONFLICT (phrase) DO UPDATE SET uses = "IraqiPhrase".uses + 1, "lastSeen" = now()`, phrase, meaning)
}

func (r *AchievementVoiceRepository) Phrases(limit int) ([]IraqiPhrase, error) {
	rows := []IraqiPhrase{}
	err := r.db.Select(&rows, `SELECT phrase, meaning, uses, "firstSeen", "lastSeen" FROM "IraqiPhrase"
		ORDER BY uses DESC, "lastSeen" DESC LIMIT $1`, limit)
	return rows, err
}
