package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 🔔 «ذكّر المدير» على طلب مواد معلّق طال عمره (11-17) ═══
// طلب ماتركس لإداري الكميات: «تذكير على طلبات المواد المعلّقة الي عمرها طال — يقل الاعتماد على
// الاتصال اليدوي». الطلب لازم عمره يومين فأكثر وبعده ما خلص، والتذكير مرة باليوم بالأكثر.

const (
	RemindMinAge   = 48 * time.Hour
	RemindCooldown = 24 * time.Hour
)

type ProcurementRemindHandler struct {
	db     *sqlx.DB
	notify func(msg string) error
}

func NewProcurementRemindHandler(db *sqlx.DB, notify func(msg string) error) *ProcurementRemindHandler {
	return &ProcurementRemindHandler{db: db, notify: notify}
}

// ErrRemind* أسباب الرفض — تنكتب للمستخدم كما هي.
var (
	ErrRemindTooNew   = errors.New("الطلب بعده جديد — التذكير بعد يومين من الطلب")
	ErrRemindTooSoon  = errors.New("انذكّر المدير بهذا الطلب خلال آخر ٢٤ ساعة")
	ErrRemindNotOpen  = errors.New("الطلب مو معلّق")
	ErrRemindNotFound = errors.New("غير موجود")
)

// RemindCheck يقرر إذا التذكير مسموح (منفصل حتى ينفحص بلا قاعدة).
func RemindCheck(status string, createdAt time.Time, remindedAt *time.Time, now time.Time) error {
	if status != "PENDING" && status != "IN_PROGRESS" {
		return ErrRemindNotOpen
	}
	if now.Sub(createdAt) < RemindMinAge {
		return ErrRemindTooNew
	}
	if remindedAt != nil && now.Sub(*remindedAt) < RemindCooldown {
		return ErrRemindTooSoon
	}
	return nil
}

// POST /api/procurement/{id}/remind
func (h *ProcurementRemindHandler) Remind(w http.ResponseWriter, r *http.Request) {
	var row struct {
		Code       string     `db:"code"`
		Status     string     `db:"status"`
		CreatedAt  time.Time  `db:"createdAt"`
		RemindedAt *time.Time `db:"remindedAt"`
	}
	err := h.db.Get(&row, `SELECT code, status::text AS status, "createdAt", "remindedAt" FROM "ProcurementRequest" WHERE id = $1`, r.PathValue("id"))
	if errors.Is(err, sql.ErrNoRows) {
		WriteError(w, http.StatusNotFound, ErrRemindNotFound.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذّر")
		return
	}
	// createdAt مخزون UTC بلا منطقة
	now := time.Now().UTC()
	created := time.Date(row.CreatedAt.Year(), row.CreatedAt.Month(), row.CreatedAt.Day(), row.CreatedAt.Hour(), row.CreatedAt.Minute(), row.CreatedAt.Second(), 0, time.UTC)
	var reminded *time.Time
	if row.RemindedAt != nil {
		t := time.Date(row.RemindedAt.Year(), row.RemindedAt.Month(), row.RemindedAt.Day(), row.RemindedAt.Hour(), row.RemindedAt.Minute(), row.RemindedAt.Second(), 0, time.UTC)
		reminded = &t
	}
	if err := RemindCheck(row.Status, created, reminded, now); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// الحجز ذرّي: ما ينرسل تذكيرين بنفس اللحظة
	res, err := h.db.Exec(`UPDATE "ProcurementRequest" SET "remindedAt" = (now() AT TIME ZONE 'UTC')
		WHERE id = $1 AND ("remindedAt" IS NULL OR "remindedAt" < (now() AT TIME ZONE 'UTC') - interval '24 hours')`, r.PathValue("id"))
	if n, _ := rowsAffected(res, err); n == 0 {
		WriteError(w, http.StatusBadRequest, ErrRemindTooSoon.Error())
		return
	}
	days := int(now.Sub(created).Hours() / 24)
	label := map[string]string{"PENDING": "بانتظار التوفير", "IN_PROGRESS": "جاري التوفير"}[row.Status]
	msg := fmt.Sprintf("🔔 تذكير من المخازن: طلب المواد %s معلّق من %d يوم (%s) — يحتاج قرارك.", row.Code, days, label)
	if err := h.notify(msg); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذّر إرسال التذكير")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func rowsAffected(res sql.Result, err error) (int64, error) {
	if err != nil || res == nil {
		return 0, err
	}
	return res.RowsAffected()
}
