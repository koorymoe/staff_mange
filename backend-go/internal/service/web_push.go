package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/jmoiron/sqlx"
)

// ═══ 🔔 إشعارات حقيقية على الموبايل (ع) 11-08 ═══
// قبل، الإشعار ينشاف بس لما الموظف يفتح الموقع — فتعيين عاجل أو طلب مدير يوصل
// متأخر. هسه كل إشعار ينكتب بجدول Notification (من أي مكان بالنظام) ينبعث لحاله
// لأجهزة صاحبه المشتركة، حتى والمتصفح مسدود.
//
// 🔒 الخصوصية: المحتوى مشفّر من الخادم للجهاز (Web Push encryption) — خدمة
// الدفع (Google/Apple/Mozilla) توصّله بس وما تكدر تقراه. ومفاتيح التشفير (VAPID)
// يولّدها الخادم أول مرة ويحفظها بالقاعدة — ما تحتاج إعداد، ومحجوبة عن الذكاء.

type WebPush struct {
	db         *sqlx.DB
	pub, priv  string
	subscriber string
	client     webpush.HTTPClient // للاختبار — فاضي = الافتراضي
}

// NewWebPush يجيب مفاتيح VAPID أو يولّدها أول مرة.
func NewWebPush(db *sqlx.DB, subscriber string) (*WebPush, error) {
	w := &WebPush{db: db, subscriber: subscriber}
	err := db.QueryRow(`SELECT "publicKey", "vapidSecret" FROM "PushVapid" WHERE id = 1`).Scan(&w.pub, &w.priv)
	if err == nil {
		return w, nil
	}
	priv, pub, gerr := webpush.GenerateVAPIDKeys()
	if gerr != nil {
		return nil, gerr
	}
	// ON CONFLICT: لو خادمين بدوا سوه، المفتاح الأول هو الي يبقى
	if _, err := db.Exec(`INSERT INTO "PushVapid" (id, "publicKey", "vapidSecret") VALUES (1, $1, $2) ON CONFLICT (id) DO NOTHING`, pub, priv); err != nil {
		return nil, err
	}
	if err := db.QueryRow(`SELECT "publicKey", "vapidSecret" FROM "PushVapid" WHERE id = 1`).Scan(&w.pub, &w.priv); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *WebPush) PublicKey() string { return w.pub }

// Subscribe يحفظ اشتراك جهاز (نفس الجهاز يتحدّث لصاحبه الحالي).
func (w *WebPush) Subscribe(empID, endpoint, p256dh, auth, ua string) error {
	if !strings.HasPrefix(endpoint, "https://") || p256dh == "" || auth == "" || len(endpoint) > 2000 {
		return errors.New("اشتراك غير صالح")
	}
	if len(ua) > 300 {
		ua = ua[:300]
	}
	_, err := w.db.Exec(`INSERT INTO "PushSubscription" (endpoint, "employeeId", p256dh, auth, "userAgent")
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (endpoint) DO UPDATE SET "employeeId" = EXCLUDED."employeeId", p256dh = EXCLUDED.p256dh, auth = EXCLUDED.auth,
		  "userAgent" = EXCLUDED."userAgent", "createdAt" = now()`, endpoint, empID, p256dh, auth, ua)
	return err
}

// Unsubscribe يشيل اشتراك جهاز (بس لصاحبه).
func (w *WebPush) Unsubscribe(empID, endpoint string) error {
	_, err := w.db.Exec(`DELETE FROM "PushSubscription" WHERE endpoint = $1 AND "employeeId" = $2`, endpoint, empID)
	return err
}

// Start يفحص كل ١٠ ثواني الإشعارات الجديدة ويبعثها.
func (w *WebPush) Start() {
	go func() {
		time.Sleep(30 * time.Second)
		for {
			if n := w.SendPending(); n > 0 {
				log.Printf("[push] انبعث %d إشعار للموبايل", n)
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

type pendingPush struct {
	ID       string `db:"id"`
	Employee string `db:"employeeId"`
	Message  string `db:"message"`
	Endpoint string `db:"endpoint"`
	P256dh   string `db:"p256dh"`
	Auth     string `db:"auth"`
}

// SendPending دورة وحدة: الإشعارات الي ما انبعثت (آخر ١٠ دقايق بس — القديم ما نزعج بيه).
func (w *WebPush) SendPending() int {
	tx, err := w.db.Beginx()
	if err != nil {
		return 0
	}
	defer tx.Rollback() //nolint:errcheck
	ids := []string{}
	// نأشّرها «انبعثت» أول (FOR UPDATE SKIP LOCKED) حتى لو أكثر من خادم ما تتكرر
	if err := tx.Select(&ids, `UPDATE "Notification" SET "pushedAt" = now() WHERE id IN (
			SELECT n.id FROM "Notification" n
			WHERE n."pushedAt" IS NULL AND n."createdAt" > now() - interval '10 minutes'
			  AND EXISTS (SELECT 1 FROM "PushSubscription" s WHERE s."employeeId" = n."employeeId")
			ORDER BY n."createdAt" LIMIT 200 FOR UPDATE SKIP LOCKED)
		RETURNING id`); err != nil || len(ids) == 0 {
		return 0
	}
	rows := []pendingPush{}
	if err := tx.Select(&rows, `SELECT n.id, n."employeeId", n.message, s.endpoint, s.p256dh, s.auth
		FROM "Notification" n JOIN "PushSubscription" s ON s."employeeId" = n."employeeId"
		WHERE n.id = ANY($1)`, ids); err != nil {
		return 0
	}
	if err := tx.Commit(); err != nil {
		return 0
	}
	sent := 0
	for _, r := range rows {
		body := r.Message
		if utf8.RuneCountInString(body) > 180 {
			body = string([]rune(body)[:180]) + "…"
		}
		payload, _ := json.Marshal(map[string]string{"title": "نظام الأماني", "body": body, "tag": r.ID})
		resp, err := webpush.SendNotification(payload, &webpush.Subscription{Endpoint: r.Endpoint, Keys: webpush.Keys{P256dh: r.P256dh, Auth: r.Auth}},
			&webpush.Options{HTTPClient: w.client, Subscriber: w.subscriber, VAPIDPublicKey: w.pub, VAPIDPrivateKey: w.priv, TTL: 3600, Urgency: webpush.UrgencyHigh})
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		switch {
		case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
			// الجهاز لغى الاشتراك — نشيله
			_, _ = w.db.Exec(`DELETE FROM "PushSubscription" WHERE endpoint = $1`, r.Endpoint)
		case resp.StatusCode < 300:
			sent++
		}
	}
	return sent
}
