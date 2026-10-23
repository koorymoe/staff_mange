package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ الميدان: موقع الشركة · الاستلام والانطلاق · تتبع المسار · خريطة المراقبة ═══
// قرار (ع) 10-14 — كل وقت ينسجّل مرتين: ضغطة الموظف وحساب النظام من الـGPS،
// حتى نقارن، ونغطي النسيان، وتنضبط أوقات الشغل.

type FieldTrackHandler struct {
	db     *sqlx.DB
	notify func(employeeID, message string)
}

// SetNotifier إشعار داخل النظام (تذكير «وصلنا»)
func (h *FieldTrackHandler) SetNotifier(f func(employeeID, message string)) { h.notify = f }

func NewFieldTrackHandler(db *sqlx.DB) *FieldTrackHandler { return &FieldTrackHandler{db: db} }

type companyLoc struct {
	Lat     float64 `db:"lat" json:"lat"`
	Lng     float64 `db:"lng" json:"lng"`
	RadiusM int     `db:"radiusM" json:"radiusM"`
}

func (h *FieldTrackHandler) company() *companyLoc {
	var c companyLoc
	if err := h.db.Get(&c, `SELECT lat, lng, "radiusM" FROM "CompanyLocation" WHERE id = 1`); err != nil {
		return nil
	}
	return &c
}

// GET /api/company-location
func (h *FieldTrackHandler) GetCompany(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{"company": h.company()})
}

// PUT /api/company-location — المالك/المدير
func (h *FieldTrackHandler) SetCompany(w http.ResponseWriter, r *http.Request) {
	var b companyLoc
	if err := DecodeJSON(r, &b); err != nil || b.Lat == 0 || b.Lng == 0 {
		WriteError(w, http.StatusBadRequest, "حدد موقع الشركة على الخريطة")
		return
	}
	if b.RadiusM < 50 || b.RadiusM > 3000 {
		b.RadiusM = 300
	}
	if _, err := h.db.Exec(`INSERT INTO "CompanyLocation" (id, lat, lng, "radiusM") VALUES (1,$1,$2,$3)
		ON CONFLICT (id) DO UPDATE SET lat = $1, lng = $2, "radiusM" = $3, "updatedAt" = now()`, b.Lat, b.Lng, b.RadiusM); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ موقع الشركة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"company": b})
}

// PUT /api/bookings/{id}/accept — «استلام المهمة» بعد التواصل ويّا الزبون
func (h *FieldTrackHandler) Accept(w http.ResponseWriter, r *http.Request) {
	if _, err := h.db.Exec(`UPDATE "Booking" SET "acceptedAt" = COALESCE("acceptedAt", now()) WHERE id = $1`, r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تسجيل الاستلام")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// classifyOrigin: منين انطلق؟ COMPANY (ضمن نطاق الشركة) · PREV_BOOKING (من عند
// زبون ثاني خلص عنده خلال آخر 3 ساعات) · OUTSIDE · UNKNOWN (بلا موقع).
func (h *FieldTrackHandler) classifyOrigin(bookingID, emp string, lat, lng *float64) string {
	if lat == nil || lng == nil {
		return "UNKNOWN"
	}
	if c := h.company(); c != nil && haversineM(*lat, *lng, c.Lat, c.Lng) <= float64(c.RadiusM) {
		return "COMPANY"
	}
	var prev []struct {
		Lat *float64 `db:"lat"`
		Lng *float64 `db:"lng"`
	}
	_ = h.db.Select(&prev, `SELECT COALESCE(b."autoArrivedLat", b."mapLatitude") AS lat, COALESCE(b."autoArrivedLng", b."mapLongitude") AS lng
		FROM "Booking" b WHERE b.id <> $1 AND b."arrivedAt" > now() - interval '12 hours'
		AND (b."projectSupervisorId" = $2 OR EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id AND a."employeeId" = $2))`, bookingID, emp)
	for _, p := range prev {
		if p.Lat != nil && p.Lng != nil && haversineM(*lat, *lng, *p.Lat, *p.Lng) <= 1500 {
			return "PREV_BOOKING"
		}
	}
	return "OUTSIDE"
}

// PUT /api/bookings/{id}/depart — «انطلاق» (منفصل عن الوصول)
func (h *FieldTrackHandler) Depart(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Lat *float64 `json:"lat"`
		Lng *float64 `json:"lng"`
	}
	_ = DecodeJSON(r, &b)
	id := r.PathValue("id")
	origin := h.classifyOrigin(id, middleware.EmployeeIDFromContext(r), b.Lat, b.Lng)
	if _, err := h.db.Exec(`UPDATE "Booking" SET "departedAt" = COALESCE("departedAt", now()),
		"departOrigin" = COALESCE("departOrigin", $2) WHERE id = $1`, id, origin); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تسجيل الانطلاق")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "origin": origin})
}

// POST /api/location-pings/batch — نقاط المسار (تنحفظ بالموبايل بلا نت وتنرسل دفعة)
func (h *FieldTrackHandler) PingBatch(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Points []struct {
			Lat       float64  `json:"lat"`
			Lng       float64  `json:"lng"`
			Accuracy  *float64 `json:"accuracy"`
			At        string   `json:"at"`
			BookingID *string  `json:"bookingId"`
		} `json:"points"`
	}
	if err := DecodeJSON(r, &b); err != nil || len(b.Points) > 2000 {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	emp, now := middleware.EmployeeIDFromContext(r), time.Now()
	tx, err := h.db.Beginx()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	defer func() { _ = tx.Rollback() }()
	saved := 0
	for _, p := range b.Points {
		at, err := time.Parse(time.RFC3339Nano, p.At)
		if err != nil || at.After(now.Add(2*time.Minute)) || at.Before(now.Add(-48*time.Hour)) || (p.Lat == 0 && p.Lng == 0) {
			continue
		}
		if p.Accuracy != nil && *p.Accuracy > 1000 {
			continue // نقطة تقديرية بلا قيمة للمسار
		}
		if _, err := tx.Exec(`INSERT INTO "LocationPing" (id, "employeeId", "bookingId", latitude, longitude, accuracy, "createdAt")
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, uuid.NewString(), emp, p.BookingID, p.Lat, p.Lng, p.Accuracy, at); err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
			return
		}
		saved++
	}
	if err := tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"saved": saved})
}

type trackBooking struct {
	ID              string     `db:"id"`
	CustomerID      string     `db:"customerId"`
	DepartedAt      *time.Time `db:"departedAt"`
	DepartOrigin    *string    `db:"departOrigin"`
	ArrivedAt       *time.Time `db:"arrivedAt"`
	CompletedAt     *time.Time `db:"completedAt"`
	MapLat          *float64   `db:"mapLatitude"`
	MapLng          *float64   `db:"mapLongitude"`
	CustLat         *float64   `db:"custLat"`
	CustLng         *float64   `db:"custLng"`
	AutoArrivedAt   *time.Time `db:"autoArrivedAt"`
	AutoCompletedAt *time.Time `db:"autoCompletedAt"`
}

func (h *FieldTrackHandler) loadPoints(bookingID string) []trackPoint {
	var rows []struct {
		Lat float64   `db:"latitude"`
		Lng float64   `db:"longitude"`
		At  time.Time `db:"createdAt"`
	}
	_ = h.db.Select(&rows, `SELECT latitude, longitude, "createdAt" FROM "LocationPing" WHERE "bookingId" = $1 ORDER BY "createdAt"`, bookingID)
	pts := make([]trackPoint, 0, len(rows))
	for _, p := range rows {
		pts = append(pts, trackPoint{p.Lat, p.Lng, p.At})
	}
	return pts
}

// analyzeBooking يحسب الأوقات التلقائية لحجز ويخزنها. يرجّع النتيجة.
func (h *FieldTrackHandler) analyzeBooking(b trackBooking) trackResult {
	pts := h.loadPoints(b.ID)
	var origin *[2]float64
	var pr struct {
		Lat *float64 `db:"lat"`
		Lng *float64 `db:"lng"`
	}
	if err := h.db.Get(&pr, `SELECT lat, lng FROM "BookingStageProof" WHERE "bookingId" = $1 AND stage = 'DEPARTED' AND lat IS NOT NULL ORDER BY "pressedAt" LIMIT 1`, b.ID); err == nil && pr.Lat != nil {
		origin = &[2]float64{*pr.Lat, *pr.Lng}
	} else if c := h.company(); c != nil && b.DepartOrigin != nil && *b.DepartOrigin == "COMPANY" {
		origin = &[2]float64{c.Lat, c.Lng}
	}
	var cust *[2]float64
	if b.CustLat != nil && b.CustLng != nil {
		cust = &[2]float64{*b.CustLat, *b.CustLng}
	} else if b.MapLat != nil && b.MapLng != nil && *b.MapLat != 0 {
		cust = &[2]float64{*b.MapLat, *b.MapLng}
	}
	res := analyzeTrack(pts, origin, cust)
	if res.Departed == nil {
		return res
	}
	var lat, lng *float64
	if res.ArrivalAt != nil {
		lat, lng = &res.ArrivalAt.Lat, &res.ArrivalAt.Lng
	}
	_, _ = h.db.Exec(`UPDATE "Booking" SET "autoDepartedAt" = $2, "autoArrivedAt" = $3, "autoCompletedAt" = $4,
		"autoArrivedLat" = $5, "autoArrivedLng" = $6 WHERE id = $1`, b.ID, res.Departed, res.Arrived, res.Completed, lat, lng)
	tx, err := h.db.Beginx()
	if err == nil {
		_, _ = tx.Exec(`DELETE FROM "BookingRouteStop" WHERE "bookingId" = $1`, b.ID)
		ins := func(kind string, s trackStop) {
			_, _ = tx.Exec(`INSERT INTO "BookingRouteStop" (id, "bookingId", kind, lat, lng, "startAt", "endAt", minutes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
				uuid.NewString(), b.ID, kind, s.Lat, s.Lng, s.Start, s.End, s.Minutes())
		}
		for _, s := range res.RoadStops {
			ins("ROAD", s)
		}
		if res.ArrivalAt != nil {
			ins("ARRIVAL", *res.ArrivalAt)
		}
		_ = tx.Commit()
	}
	if res.Arrived != nil {
		// الموظف نسى يضغط «وصلنا» → النظام يحسبله (ويبين «تلقائي» بالدليل)
		if b.ArrivedAt == nil {
			if r, err := h.db.Exec(`UPDATE "Booking" SET "arrivedAt" = $2 WHERE id = $1 AND "arrivedAt" IS NULL`, b.ID, *res.Arrived); err == nil {
				if n, _ := r.RowsAffected(); n > 0 {
					_, _ = h.db.Exec(`INSERT INTO "BookingStageProof" (id, "bookingId", stage, "pressedAt", offline, lat, lng, verdict)
						VALUES ($1,$2,'ARRIVED',$3,false,$4,$5,'AUTO')`, uuid.NewString(), b.ID, *res.Arrived, lat, lng)
				}
			}
		}
		// موقع الزبون: إذا ما محدد رسمياً (من التطبيق)، ناخذ مكان التوقف الطويل
		_, _ = h.db.Exec(`UPDATE "Customer" SET lat = $2, lng = $3, "locationSource" = 'AUTO'
			WHERE id = $1 AND COALESCE("locationSource", '') <> 'APP'`, b.CustomerID, *lat, *lng)
	}
	return res
}

const trackBookingSQL = `SELECT b.id, b."customerId", b."departedAt", b."departOrigin", b."arrivedAt", b."completedAt",
	b."mapLatitude", b."mapLongitude", c.lat AS "custLat", c.lng AS "custLng", b."autoArrivedAt", b."autoCompletedAt"
	FROM "Booking" b LEFT JOIN "Customer" c ON c.id = b."customerId" AND c."locationSource" = 'APP'`

// RunLoop كل دقيقتين: الحجوزات الي انطلقت خلال آخر يوم وما انحسب إنجازها التلقائي.
func (h *FieldTrackHandler) RunLoop() {
	for {
		time.Sleep(2 * time.Minute)
		var rows []trackBooking
		if err := h.db.Select(&rows, trackBookingSQL+` WHERE (b."departedAt" > now() - interval '24 hours'
			OR EXISTS (SELECT 1 FROM "LocationPing" p WHERE p."bookingId" = b.id AND p."createdAt" > now() - interval '24 hours'))
			AND b."autoCompletedAt" IS NULL`); err != nil {
			log.Printf("[field-track] %v", err)
			continue
		}
		for _, b := range rows {
			h.analyzeBooking(b)
			h.remindArrival(b)
		}
	}
}

// GET /api/bookings/{id}/track — المسار والتوقفات والمقارنة (يدوي ↔ تلقائي)
func (h *FieldTrackHandler) Track(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var b trackBooking
	if err := h.db.Get(&b, trackBookingSQL+` WHERE b.id = $1`, id); err != nil {
		WriteError(w, http.StatusNotFound, "الحجز غير موجود")
		return
	}
	h.analyzeBooking(b)
	var times struct {
		AcceptedAt      *time.Time `db:"acceptedAt" json:"acceptedAt"`
		MaterialsAt     *time.Time `db:"materialsReadyAt" json:"materialsReadyAt"`
		DepartedAt      *time.Time `db:"departedAt" json:"departedAt"`
		DepartOrigin    *string    `db:"departOrigin" json:"departOrigin"`
		ArrivedAt       *time.Time `db:"arrivedAt" json:"arrivedAt"`
		StartedAt       *time.Time `db:"startedAt" json:"startedAt"`
		CompletedAt     *time.Time `db:"completedAt" json:"completedAt"`
		AutoDepartedAt  *time.Time `db:"autoDepartedAt" json:"autoDepartedAt"`
		AutoArrivedAt   *time.Time `db:"autoArrivedAt" json:"autoArrivedAt"`
		AutoCompletedAt *time.Time `db:"autoCompletedAt" json:"autoCompletedAt"`
		Lat             *float64   `db:"mapLatitude" json:"lat"`
		Lng             *float64   `db:"mapLongitude" json:"lng"`
	}
	_ = h.db.Get(&times, `SELECT "acceptedAt", "materialsReadyAt", "departedAt", "departOrigin", "arrivedAt", "startedAt", "completedAt",
		"autoDepartedAt", "autoArrivedAt", "autoCompletedAt", "mapLatitude", "mapLongitude" FROM "Booking" WHERE id = $1`, id)
	type stopRow struct {
		Kind    string    `db:"kind" json:"kind"`
		Lat     float64   `db:"lat" json:"lat"`
		Lng     float64   `db:"lng" json:"lng"`
		StartAt time.Time `db:"startAt" json:"startAt"`
		EndAt   time.Time `db:"endAt" json:"endAt"`
		Minutes int       `db:"minutes" json:"minutes"`
	}
	stops := []stopRow{}
	_ = h.db.Select(&stops, `SELECT kind, lat, lng, "startAt", "endAt", minutes FROM "BookingRouteStop" WHERE "bookingId" = $1 ORDER BY "startAt"`, id)
	type proofRow struct {
		Stage     string    `db:"stage" json:"stage"`
		PressedAt time.Time `db:"pressedAt" json:"pressedAt"`
		Offline   bool      `db:"offline" json:"offline"`
		Verdict   string    `db:"verdict" json:"verdict"`
		DistanceM *float64  `db:"distanceM" json:"distanceM"`
	}
	proofs := []proofRow{}
	_ = h.db.Select(&proofs, `SELECT stage, "pressedAt", offline, verdict, "distanceM" FROM "BookingStageProof" WHERE "bookingId" = $1 ORDER BY "pressedAt"`, id)
	pts := h.loadPoints(id)
	path := make([][2]float64, 0, len(pts))
	for _, p := range pts {
		path = append(path, [2]float64{p.Lat, p.Lng})
	}
	WriteJSON(w, http.StatusOK, map[string]any{"times": times, "stops": stops, "proofs": proofs, "path": path, "company": h.company()})
}

// GET /api/field-map?from=YYYY-MM-DD&to=YYYY-MM-DD — خريطة الميدان للمالك والمدير والمراقب
func (h *FieldTrackHandler) FieldMap(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from, to := q.Get("from"), q.Get("to")
	if from == "" {
		from = time.Now().Format("2006-01-02")
	}
	if to == "" {
		to = from
	}
	type bookingRow struct {
		ID              string     `db:"id" json:"id"`
		Code            string     `db:"code" json:"code"`
		Customer        *string    `db:"customer" json:"customer"`
		Status          string     `db:"status" json:"status"`
		Lat             *float64   `db:"lat" json:"lat"`
		Lng             *float64   `db:"lng" json:"lng"`
		DepartedAt      *time.Time `db:"departedAt" json:"departedAt"`
		DepartOrigin    *string    `db:"departOrigin" json:"departOrigin"`
		ArrivedAt       *time.Time `db:"arrivedAt" json:"arrivedAt"`
		CompletedAt     *time.Time `db:"completedAt" json:"completedAt"`
		AutoArrivedAt   *time.Time `db:"autoArrivedAt" json:"autoArrivedAt"`
		AutoCompletedAt *time.Time `db:"autoCompletedAt" json:"autoCompletedAt"`
		Team            *string    `db:"team" json:"team"`
	}
	bookings := []bookingRow{}
	if err := h.db.Select(&bookings, `SELECT b.id, b.code::text AS code, c.name AS customer, b.status::text AS status,
		COALESCE(cl.lat, b."mapLatitude") AS lat, COALESCE(cl.lng, b."mapLongitude") AS lng,
		b."departedAt", b."departOrigin", b."arrivedAt", b."completedAt", b."autoArrivedAt", b."autoCompletedAt",
		(SELECT string_agg(e.name, '، ') FROM "BookingAssignment" a JOIN "Employee" e ON e.id = a."employeeId" WHERE a."bookingId" = b.id) AS team
		FROM "Booking" b LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Customer" cl ON cl.id = b."customerId" AND cl."locationSource" = 'APP'
		WHERE b."archivedAt" IS NULL AND baghdad_date(COALESCE(b."departedAt", b."arrivedAt", b."scheduledAt", b."createdAt")) BETWEEN $1::date AND $2::date
		ORDER BY COALESCE(b."departedAt", b."scheduledAt")`, from, to); err != nil {
		log.Printf("[field-map] %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب خريطة الميدان")
		return
	}
	type empRow struct {
		EmployeeID string    `db:"employeeId" json:"employeeId"`
		Name       string    `db:"name" json:"name"`
		Lat        float64   `db:"latitude" json:"lat"`
		Lng        float64   `db:"longitude" json:"lng"`
		At         time.Time `db:"createdAt" json:"at"`
		BookingID  *string   `db:"bookingId" json:"bookingId"`
	}
	latest := []empRow{}
	_ = h.db.Select(&latest, `SELECT DISTINCT ON (p."employeeId") p."employeeId", e.name, p.latitude, p.longitude, p."createdAt", p."bookingId"
		FROM "LocationPing" p JOIN "Employee" e ON e.id = p."employeeId"
		WHERE baghdad_date(p."createdAt") BETWEEN $1::date AND $2::date ORDER BY p."employeeId", p."createdAt" DESC`, from, to)
	// نقاط الخريطة الحرارية: التوقفات بوزن مدتها + عيّنة من نقاط المسار
	heat := [][3]float64{}
	var stops []struct {
		Lat float64 `db:"lat"`
		Lng float64 `db:"lng"`
		Min int     `db:"minutes"`
	}
	_ = h.db.Select(&stops, `SELECT s.lat, s.lng, s.minutes FROM "BookingRouteStop" s WHERE baghdad_date(s."startAt") BETWEEN $1::date AND $2::date`, from, to)
	for _, s := range stops {
		heat = append(heat, [3]float64{s.Lat, s.Lng, float64(s.Min)})
	}
	var pts []struct {
		Lat float64 `db:"latitude"`
		Lng float64 `db:"longitude"`
	}
	_ = h.db.Select(&pts, `SELECT latitude, longitude FROM "LocationPing" WHERE baghdad_date("createdAt") BETWEEN $1::date AND $2::date
		ORDER BY random() LIMIT 5000`, from, to)
	for _, p := range pts {
		heat = append(heat, [3]float64{p.Lat, p.Lng, 1})
	}
	WriteJSON(w, http.StatusOK, map[string]any{"company": h.company(), "bookings": bookings, "employees": latest, "heat": heat})
}

// remindArrival (ع) 10-15: النظام شاف الموظف واكف ١٠ دقايق أو أكثر قرب الزبون وهو
// ما ضاغط «وصلنا» — نذكّره مرة وحدة. (بعد ٣٠ دقيقة النظام يسجّلها بداله.)
func (h *FieldTrackHandler) remindArrival(b trackBooking) {
	if h.notify == nil || b.DepartedAt == nil {
		return
	}
	var st struct {
		ArrivedAt  *time.Time `db:"arrivedAt"`
		RemindedAt *time.Time `db:"arriveRemindedAt"`
		Code       string     `db:"code"`
		Driver     *string    `db:"driver"`
	}
	if err := h.db.Get(&st, `SELECT b."arrivedAt", b."arriveRemindedAt", b.code::text AS code,
		COALESCE((SELECT p."employeeId" FROM "BookingStageProof" p WHERE p."bookingId" = b.id AND p.stage = 'DEPARTED' AND p."employeeId" IS NOT NULL ORDER BY p."pressedAt" LIMIT 1), b."projectSupervisorId") AS driver
		FROM "Booking" b WHERE b.id = $1`, b.ID); err != nil || st.ArrivedAt != nil || st.RemindedAt != nil || st.Driver == nil {
		return
	}
	s := currentStop(h.loadPoints(b.ID))
	if s == nil || s.End.Sub(s.Start) < 10*time.Minute || time.Since(s.End) > 10*time.Minute {
		return
	}
	var cust *[2]float64
	if b.CustLat != nil && b.CustLng != nil {
		cust = &[2]float64{*b.CustLat, *b.CustLng}
	} else if b.MapLat != nil && b.MapLng != nil && *b.MapLat != 0 {
		cust = &[2]float64{*b.MapLat, *b.MapLng}
	}
	if cust != nil && haversineM(cust[0], cust[1], s.Lat, s.Lng) > customerNearM {
		return
	}
	if r, err := h.db.Exec(`UPDATE "Booking" SET "arriveRemindedAt" = now() WHERE id = $1 AND "arriveRemindedAt" IS NULL`, b.ID); err == nil {
		if n, _ := r.RowsAffected(); n > 0 {
			h.notify(*st.Driver, fmt.Sprintf("📍 الحجز %s: النظام شافك عند الزبون من %d دقيقة — دوس «وصلنا» من «مهامي»", st.Code, s.Minutes()))
		}
	}
}
