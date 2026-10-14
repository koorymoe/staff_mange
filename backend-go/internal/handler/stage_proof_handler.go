package handler

import (
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ مراحل الحجز بوقتها الحقيقي — قرار (ع) 10-14 ═══
//
// الليدر يدوس «وصلنا / بدأنا / تم الإنجاز» حتى لو ماكو نت: الموبايل يحفظ
// الضغطة بوقتها وموقعها ويرسلها من يرجع النت. هنا:
//   - نسجّل الدليل (وقت الضغطة، الموقع، دقته، البعد عن الزبون).
//   - إذا الضغطة انحفظت بلا نت، نصحّح وقت المرحلة بالحجز لوقت الضغطة
//     الحقيقي (مو وقت وصول الطلب) — بشرط يكون منطقي ومرتب.
//   - الحكم (verdict): OK | FAR (أبعد من كيلو عن الزبون بدقة كافية)
//     | BATCH (٣ مراحل أو أكثر من نفس الموظف خلال ٥ دقايق) — دليل مو منع.

type StageProofHandler struct{ db *sqlx.DB }

func NewStageProofHandler(db *sqlx.DB) *StageProofHandler { return &StageProofHandler{db: db} }

var stageColumn = map[string]string{"ARRIVED": "arrivedAt", "STARTED": "startedAt", "COMPLETED": "completedAt"}

func haversineM(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371000.0
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat, dLng := toRad(lat2-lat1), toRad(lng2-lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * r * math.Asin(math.Sqrt(a))
}

// POST /api/bookings/{id}/stage-proof
func (h *StageProofHandler) Record(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Stage     string   `json:"stage"`
		PressedAt string   `json:"pressedAt"`
		Offline   bool     `json:"offline"`
		Lat       *float64 `json:"lat"`
		Lng       *float64 `json:"lng"`
		Accuracy  *float64 `json:"accuracy"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	col, ok := stageColumn[b.Stage]
	if !ok {
		WriteError(w, http.StatusBadRequest, "مرحلة غير معروفة")
		return
	}
	now := time.Now()
	pressed, err := time.Parse(time.RFC3339Nano, b.PressedAt)
	// وقت الضغطة لازم يكون بآخر ٢٤ ساعة ومو بالمستقبل (هامش دقيقتين لفرق الساعات)
	if err != nil || pressed.After(now.Add(2*time.Minute)) || pressed.Before(now.Add(-24*time.Hour)) {
		pressed = now
	}
	if pressed.After(now) {
		pressed = now
	}
	bookingID, emp := r.PathValue("id"), middleware.EmployeeIDFromContext(r)

	var bk struct {
		Lat *float64 `db:"mapLatitude"`
		Lng *float64 `db:"mapLongitude"`
	}
	if err := h.db.Get(&bk, `SELECT "mapLatitude", "mapLongitude" FROM "Booking" WHERE id = $1`, bookingID); err != nil {
		WriteError(w, http.StatusNotFound, "الحجز غير موجود")
		return
	}
	verdict := "OK"
	var dist *float64
	if b.Lat != nil && b.Lng != nil && bk.Lat != nil && bk.Lng != nil && *bk.Lat != 0 {
		d := haversineM(*b.Lat, *b.Lng, *bk.Lat, *bk.Lng)
		dist = &d
		// الدقة الضعيفة (أكثر من ٥٠٠ م) ما نحكم بيها — GPS داخل البيوت يغلط
		if d > 1000 && (b.Accuracy == nil || *b.Accuracy <= 500) {
			verdict = "FAR"
		}
	}
	// تسجيل جماعي: هالضغطة + ضغطتين أو أكثر من نفس الموظف خلال ٥ دقايق
	var near int
	_ = h.db.Get(&near, `SELECT count(*) FROM "BookingStageProof" WHERE "employeeId" = $1 AND "bookingId" <> $2
		AND "pressedAt" BETWEEN $3::timestamp - interval '5 minutes' AND $3::timestamp + interval '5 minutes'`, emp, bookingID, pressed)
	if near >= 2 {
		verdict = "BATCH"
		_, _ = h.db.Exec(`UPDATE "BookingStageProof" SET verdict = 'BATCH' WHERE "employeeId" = $1 AND verdict = 'OK'
			AND "pressedAt" BETWEEN $2::timestamp - interval '5 minutes' AND $2::timestamp + interval '5 minutes'`, emp, pressed)
	}
	if _, err := h.db.Exec(`INSERT INTO "BookingStageProof" (id, "bookingId", stage, "employeeId", "pressedAt", offline, lat, lng, accuracy, "distanceM", verdict)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, uuid.NewString(), bookingID, b.Stage, emp, pressed, b.Offline, b.Lat, b.Lng, b.Accuracy, dist, verdict); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الدليل")
		return
	}
	// الضغطة الأوفلاين: وقت المرحلة = وقت الضغطة الحقيقي، بشرط الترتيب
	// (الوصول ≤ البدء ≤ الإنجاز) وما يتأخر عن الوقت المسجّل.
	if b.Offline {
		_, _ = h.db.Exec(`UPDATE "Booking" SET "`+col+`" = $2
			WHERE id = $1 AND "`+col+`" IS NOT NULL AND $2 < "`+col+`"
			  AND ($3 <> 'STARTED' OR "arrivedAt" IS NULL OR "arrivedAt" <= $2)
			  AND ($3 <> 'COMPLETED' OR "startedAt" IS NULL OR "startedAt" <= $2)
			  AND ($3 <> 'ARRIVED' OR "startedAt" IS NULL OR $2 <= "startedAt")`, bookingID, pressed, b.Stage)
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "verdict": verdict, "distanceM": dist, "pressedAt": pressed})
}
