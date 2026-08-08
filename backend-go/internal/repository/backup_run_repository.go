package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// BackupRunRepository سجل النسخ الاحتياطية — للمالك وحده.
//
// السكربت على السيرفر يكتب صف بكل تشغيل (ناجح أو فاشل)، وهنا نقراه
// ونحكم عليه. الحكم مو مجرد «آخر نسخة نجحت»: نسخة نجحت بس ما طلعت
// برّا السيرفر ما تحميك من ضياع السيرفر، ونسخة عمرها أربع أيام معناها
// الجدولة واقفة حتى لو آخر صف مكتوب عليه ✓.
type BackupRunRepository struct {
	db *sqlx.DB
}

func NewBackupRunRepository(db *sqlx.DB) *BackupRunRepository {
	return &BackupRunRepository{db: db}
}

// BackupHealth خلاصة جاهزة للعرض — الحكم ينحسب هنا بالسيرفر مو
// بالواجهة، حتى يبقى نفس الحكم لو انقرا من أي مكان.
type BackupHealth struct {
	Last              *model.BackupRun `json:"last"`
	LastSuccess       *model.BackupRun `json:"lastSuccess"`
	HoursSinceSuccess *float64         `json:"hoursSinceSuccess"`
	ConsecutiveFails  int              `json:"consecutiveFails"`
	Success7d         int              `json:"success7d"`
	Failed7d          int              `json:"failed7d"`
	OffsiteOK         bool             `json:"offsiteOK"`
	Encrypted         bool             `json:"encrypted"`
	Status            string           `json:"status"` // OK | WARN | CRITICAL | UNKNOWN
	Message           string           `json:"message"`
	Recommendations   []string         `json:"recommendations"`
	TotalRunsRecorded int              `json:"totalRunsRecorded"`
}

func (r *BackupRunRepository) List(limit int) ([]model.BackupRun, error) {
	if limit <= 0 || limit > 200 {
		limit = 30
	}
	rows := []model.BackupRun{}
	err := r.db.Select(&rows, `
		SELECT * FROM "BackupRun" ORDER BY "startedAt" DESC LIMIT $1`, limit)
	return rows, err
}

func (r *BackupRunRepository) Health() (*BackupHealth, error) {
	h := &BackupHealth{Status: "UNKNOWN", Recommendations: []string{}}

	var last model.BackupRun
	err := r.db.Get(&last, `SELECT * FROM "BackupRun" ORDER BY "startedAt" DESC LIMIT 1`)
	if errors.Is(err, sql.ErrNoRows) {
		h.Message = "ماكو ولا نسخة مسجّلة لحد الآن — شغّل ./setup-backups.sh على السيرفر"
		h.Recommendations = append(h.Recommendations, "شغّل ./setup-backups.sh مرة وحدة حتى تنجدول النسخة اليومية")
		return h, nil
	}
	if err != nil {
		return nil, err
	}
	h.Last = &last

	var ok model.BackupRun
	err = r.db.Get(&ok, `SELECT * FROM "BackupRun" WHERE ok ORDER BY "startedAt" DESC LIMIT 1`)
	switch {
	case err == nil:
		h.LastSuccess = &ok
		hours := time.Since(ok.StartedAt).Hours()
		h.HoursSinceSuccess = &hours
		h.OffsiteOK = ok.Offsite
		h.Encrypted = ok.Encrypted
	case errors.Is(err, sql.ErrNoRows):
		// ماكو ولا نسخة ناجحة أبداً — أخطر حالة ممكنة
	default:
		return nil, err
	}

	if err := r.db.Get(&h.Success7d, `
		SELECT COUNT(*) FROM "BackupRun"
		WHERE ok AND "startedAt" > NOW() - INTERVAL '7 days'`); err != nil {
		return nil, err
	}
	if err := r.db.Get(&h.Failed7d, `
		SELECT COUNT(*) FROM "BackupRun"
		WHERE NOT ok AND "startedAt" > NOW() - INTERVAL '7 days'`); err != nil {
		return nil, err
	}
	if err := r.db.Get(&h.TotalRunsRecorded, `SELECT COUNT(*) FROM "BackupRun"`); err != nil {
		return nil, err
	}

	// كم فشل متتالي من آخر نسخة للورا — نوكف بأول نجاح.
	recent := []bool{}
	if err := r.db.Select(&recent, `
		SELECT ok FROM "BackupRun" ORDER BY "startedAt" DESC LIMIT 20`); err != nil {
		return nil, err
	}
	for _, v := range recent {
		if v {
			break
		}
		h.ConsecutiveFails++
	}

	switch {
	case h.LastSuccess == nil:
		h.Status = "CRITICAL"
		h.Message = "ماكو ولا نسخة ناجحة — النظام بلا حماية نهائياً"
	case *h.HoursSinceSuccess > 72:
		h.Status = "CRITICAL"
		h.Message = "آخر نسخة ناجحة من أكثر من ٣ أيام — الجدولة على الأغلب واقفة"
	case *h.HoursSinceSuccess > 30:
		h.Status = "WARN"
		h.Message = "فاتت نسخة يومية — آخر نجاح صار قبل أكثر من يوم"
	case h.ConsecutiveFails > 0:
		h.Status = "WARN"
		h.Message = "آخر محاولة نسخ فشلت (بس اكو نسخة ناجحة قريبة)"
	case !h.OffsiteOK:
		h.Status = "WARN"
		h.Message = "النسخ شغالة بس محفوظة على نفس السيرفر بس — لو ضاع السيرفر تروح وياه"
	default:
		h.Status = "OK"
		h.Message = "النسخ شغالة ومحفوظة خارج السيرفر"
	}

	if !h.OffsiteOK {
		h.Recommendations = append(h.Recommendations,
			"فعّل وجهة خارج السيرفر (R2 أو SSH) من ملف .env — بدونها النسخة ما تحميك من ضياع السيرفر")
	}
	if h.LastSuccess != nil && !h.Encrypted {
		h.Recommendations = append(h.Recommendations,
			"ضيف BACKUP_PASSPHRASE بـ.env — النسخة فيها كلمات سر وبيانات زبائن وهي حالياً غير مشفّرة")
	}
	if h.ConsecutiveFails >= 2 {
		h.Recommendations = append(h.Recommendations,
			"فشل متكرر — ادخل السيرفر وشغّل ./backup-db.sh يدوياً وشوف الخطأ")
	}
	if h.LastSuccess != nil && h.LastSuccess.TableCount > 0 && h.LastSuccess.TableCount < 100 {
		h.Recommendations = append(h.Recommendations,
			"عدد الجداول بآخر نسخة أقل من المتوقع — تأكد إن النسخة كاملة")
	}
	return h, nil
}
