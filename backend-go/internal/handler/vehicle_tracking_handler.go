package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
)

// ═══ متابعة السيارات الشهرية — نفس ملف إكسل (ع) 10-09 ═══
// الفترة من يوم ١٧ لحد ١٦ بالشهر الجاي. كل يوم جدول لكل السيارات، وإحصائيات
// كل سيارة بنفس أعمدة شيت «الإحصائيات». الغسل: «تم» = 4، «لم يتم» = 0.
type VehicleTrackingHandler struct{ db *sqlx.DB }

func NewVehicleTrackingHandler(db *sqlx.DB) *VehicleTrackingHandler {
	return &VehicleTrackingHandler{db: db}
}

// أوزان البنود العشرة (بدون الغسل) — نفس الملف. الغسل وزنه 8% بالشهري.
var trackItems = []struct {
	Key    string
	Label  string
	Weight float64
	get    func(*model.VehicleDailyRating) *int
}{
	{"exteriorClean", "نظافة الهيكل الخارجي", 0.07, func(r *model.VehicleDailyRating) *int { return r.ExteriorClean }},
	{"exteriorCondition", "حالة الهيكل الخارجي", 0.10, func(r *model.VehicleDailyRating) *int { return r.ExteriorCondition }},
	{"tireCondition", "حالة الإطارات", 0.10, func(r *model.VehicleDailyRating) *int { return r.TireCondition }},
	{"glassClean", "تنظيف الزجاج", 0.05, func(r *model.VehicleDailyRating) *int { return r.GlassClean }},
	{"lightsCondition", "حالة اللايتات", 0.05, func(r *model.VehicleDailyRating) *int { return r.LightsCondition }},
	{"technicalFaults", "الأعطال الفنية", 0.25, func(r *model.VehicleDailyRating) *int { return r.TechnicalFaults }},
	{"interiorClean", "التنظيف الداخلي", 0.10, func(r *model.VehicleDailyRating) *int { return r.InteriorClean }},
	{"seatsCondition", "حالة الكراسي", 0.07, func(r *model.VehicleDailyRating) *int { return r.SeatsCondition }},
	{"interiorDirt", "الأوساخ الداخلية", 0.08, func(r *model.VehicleDailyRating) *int { return r.InteriorDirt }},
	{"smell", "الرائحة", 0.05, func(r *model.VehicleDailyRating) *int { return r.Smell }},
}

const washWeight = 0.08

// DailyScore النسبة الموزونة لليوم (بدون الغسل) — عمود «التقييم الموزون %».
func DailyScore(r *model.VehicleDailyRating) *float64 {
	var num, den float64
	for _, it := range trackItems {
		if v := it.get(r); v != nil {
			num += float64(*v) * it.Weight
			den += 4 * it.Weight
		}
	}
	if den == 0 {
		return nil
	}
	v := num / den * 100
	return &v
}

// VehicleTrackStats صف بشيت «الإحصائيات».
type VehicleTrackStats struct {
	VehicleID   string              `json:"vehicleId"`
	VehicleName string              `json:"vehicleName"`
	Completed   int                 `json:"completed"`
	RawSum      int                 `json:"rawSum"`
	RawMax      int                 `json:"rawMax"`
	RawPct      *float64            `json:"rawPct"`
	WashPct     *float64            `json:"washPct"`
	Items       map[string]*float64 `json:"items"`
	Weighted    *float64            `json:"weighted"`
	Grade       string              `json:"grade"`
	WashDone    int                 `json:"washDone"`
	WashNot     int                 `json:"washNot"`
}

func gradeOf(p *float64) string {
	if p == nil {
		return ""
	}
	switch {
	case *p >= 90:
		return "ممتاز"
	case *p >= 80:
		return "جيد جداً"
	case *p >= 70:
		return "جيد"
	case *p >= 60:
		return "مقبول"
	}
	return "ضعيف"
}

func pctPtr(num, den float64) *float64 {
	if den == 0 {
		return nil
	}
	v := num / den * 100
	return &v
}

// BuildTrackStats نفس معادلات شيت الإحصائيات.
func BuildTrackStats(id, name string, rows []model.VehicleDailyRating) VehicleTrackStats {
	st := VehicleTrackStats{VehicleID: id, VehicleName: name, Items: map[string]*float64{}}
	var washSum, washN float64
	itemSum := map[string]float64{}
	itemN := map[string]float64{}
	for i := range rows {
		r := &rows[i]
		if DailyScore(r) != nil {
			st.Completed++
		}
		if r.Wash != nil {
			washSum += float64(*r.Wash) / 4
			washN++
			if *r.Wash >= 2 {
				st.WashDone++
			} else {
				st.WashNot++
			}
		}
		for _, it := range trackItems {
			if v := it.get(r); v != nil {
				st.RawSum += *v
				st.RawMax += 4
				itemSum[it.Key] += float64(*v)
				itemN[it.Key]++
			}
		}
	}
	st.RawPct = pctPtr(float64(st.RawSum), float64(st.RawMax))
	st.WashPct = pctPtr(washSum, washN)
	var num, den float64
	if st.WashPct != nil {
		num += *st.WashPct * washWeight
		den += washWeight
	}
	for _, it := range trackItems {
		p := pctPtr(itemSum[it.Key], itemN[it.Key]*4)
		st.Items[it.Key] = p
		if p != nil {
			num += *p * it.Weight
			den += it.Weight
		}
	}
	if den > 0 {
		v := num / den
		st.Weighted = &v
	}
	st.Grade = gradeOf(st.Weighted)
	return st
}

// periodRange «2026-09» → من 2026-09-17 لحد 2026-10-16.
func periodRange(p string) (time.Time, time.Time) {
	t, err := time.Parse("2006-01", p)
	if err != nil {
		now := time.Now().In(time.FixedZone("Baghdad", 3*3600))
		if now.Day() >= 17 {
			t = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		} else {
			t = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
		}
	}
	from := time.Date(t.Year(), t.Month(), 17, 0, 0, 0, 0, time.UTC)
	return from, from.AddDate(0, 1, -1)
}

type trackRow struct {
	model.VehicleDailyRating
	DateKey string   `json:"date"`
	Score   *float64 `json:"score"`
}

// GET /api/vehicles/tracking?period=YYYY-MM
func (h *VehicleTrackingHandler) Get(w http.ResponseWriter, r *http.Request) {
	from, to := periodRange(r.URL.Query().Get("period"))
	vehicles := []struct {
		ID   string `db:"id" json:"id"`
		Name string `db:"name" json:"name"`
		// Temp = سيارة انضافت من الإكسل برقم مؤقت — يگدر يدمجها ويا سيارة النظام
		Temp bool `db:"temp" json:"temp"`
	}{}
	if err := h.db.Select(&vehicles, `SELECT id, name, "plateNumber" LIKE 'بلا-رقم-%' AS temp FROM "Vehicle" WHERE "isActive" ORDER BY "createdAt", name`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب السيارات")
		return
	}
	ratings := []model.VehicleDailyRating{}
	if err := h.db.Select(&ratings, `SELECT r.* FROM "VehicleDailyRating" r JOIN "Vehicle" v ON v.id = r."vehicleId" AND v."isActive"
		WHERE r."ratedDate" BETWEEN $1::date AND $2::date ORDER BY r."ratedDate", r."createdAt"`,
		from.Format("2006-01-02"), to.Format("2006-01-02")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التقييمات")
		return
	}
	byVehicle := map[string][]model.VehicleDailyRating{}
	rows := make([]trackRow, 0, len(ratings))
	for _, rt := range ratings {
		byVehicle[rt.VehicleID] = append(byVehicle[rt.VehicleID], rt)
		rows = append(rows, trackRow{VehicleDailyRating: rt, DateKey: rt.RatedDate.Format("2006-01-02"), Score: DailyScore(&rt)})
	}
	stats := make([]VehicleTrackStats, 0, len(vehicles))
	for _, v := range vehicles {
		stats = append(stats, BuildTrackStats(v.ID, v.Name, byVehicle[v.ID]))
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"),
		"vehicles": vehicles, "ratings": rows, "stats": stats,
	})
}

// POST /api/vehicles/tracking/merge {fromId, toId}
// (ع) 10-09: سيارات الإكسل الي اسمها غير عن اسمها بالنظام انضافت مرتين. أبو
// الكميات يختار «هاي نفس سيارة …» فتنتقل تقييماتها لسيارة النظام (اليوم
// المسجّل أصلاً ما يتكرر) والمؤقتة تنطفي (isActive=false) — ما تنمسح.
// بس السيارة المؤقتة (رقم «بلا-رقم-») تنقبل كمصدر.
func (h *VehicleTrackingHandler) Merge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromID string `json:"fromId"`
		ToID   string `json:"toId"`
	}
	if err := DecodeJSON(r, &req); err != nil || req.FromID == "" || req.ToID == "" || req.FromID == req.ToID {
		WriteError(w, http.StatusBadRequest, "اختار السيارتين")
		return
	}
	tx, err := h.db.Beginx()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الدمج")
		return
	}
	defer tx.Rollback()
	var ok bool
	if err := tx.Get(&ok, `SELECT EXISTS(SELECT 1 FROM "Vehicle" WHERE id=$1 AND "plateNumber" LIKE 'بلا-رقم-%')
		AND EXISTS(SELECT 1 FROM "Vehicle" WHERE id=$2 AND "isActive")`, req.FromID, req.ToID); err != nil || !ok {
		WriteError(w, http.StatusBadRequest, "الدمج بس للسيارات المضافة من الإكسل")
		return
	}
	if _, err := tx.Exec(`UPDATE "VehicleDailyRating" r SET "vehicleId"=$2 WHERE r."vehicleId"=$1
		AND NOT EXISTS (SELECT 1 FROM "VehicleDailyRating" x WHERE x."vehicleId"=$2 AND x."ratedDate"=r."ratedDate")`, req.FromID, req.ToID); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر نقل التقييمات")
		return
	}
	if _, err := tx.Exec(`UPDATE "Vehicle" SET "isActive"=false WHERE id=$1`, req.FromID); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الدمج")
		return
	}
	if err := tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الدمج")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type trackDayRow struct {
	VehicleID         string  `json:"vehicleId"`
	Wash              *int    `json:"wash"`
	ExteriorClean     *int    `json:"exteriorClean"`
	ExteriorCondition *int    `json:"exteriorCondition"`
	TireCondition     *int    `json:"tireCondition"`
	GlassClean        *int    `json:"glassClean"`
	LightsCondition   *int    `json:"lightsCondition"`
	TechnicalFaults   *int    `json:"technicalFaults"`
	FaultDescription  *string `json:"faultDescription"`
	InteriorClean     *int    `json:"interiorClean"`
	SeatsCondition    *int    `json:"seatsCondition"`
	InteriorDirt      *int    `json:"interiorDirt"`
	Smell             *int    `json:"smell"`
}

func validScore(p *int) bool { return p == nil || (*p >= 0 && *p <= 4) }

// PUT /api/vehicles/tracking/day {date, rows[]} — يحفظ اليوم كله لكل السيارات.
func (h *VehicleTrackingHandler) SaveDay(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Date string        `json:"date"`
		Rows []trackDayRow `json:"rows"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if _, err := time.Parse("2006-01-02", body.Date); err != nil {
		WriteError(w, http.StatusBadRequest, "حدد اليوم")
		return
	}
	me := middleware.EmployeeIDFromContext(r)
	tx, err := h.db.Beginx()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	defer func() { _ = tx.Rollback() }()
	for _, x := range body.Rows {
		for _, p := range []*int{x.Wash, x.ExteriorClean, x.ExteriorCondition, x.TireCondition, x.GlassClean, x.LightsCondition,
			x.TechnicalFaults, x.InteriorClean, x.SeatsCondition, x.InteriorDirt, x.Smell} {
			if !validScore(p) {
				WriteError(w, http.StatusBadRequest, "الدرجة لازم من 0 لـ4")
				return
			}
		}
		if x.FaultDescription != nil {
			t := strings.TrimSpace(*x.FaultDescription)
			x.FaultDescription = &t
			if t == "" {
				x.FaultDescription = nil
			}
		}
		args := []any{x.VehicleID, body.Date, x.Wash, x.ExteriorClean, x.ExteriorCondition, x.TireCondition, x.GlassClean,
			x.LightsCondition, x.TechnicalFaults, x.FaultDescription, x.InteriorClean, x.SeatsCondition, x.InteriorDirt, x.Smell, me}
		res, err := tx.Exec(`UPDATE "VehicleDailyRating" SET wash=$3, "exteriorClean"=$4, "exteriorCondition"=$5, "tireCondition"=$6,
			"glassClean"=$7, "lightsCondition"=$8, "technicalFaults"=$9, "faultDescription"=$10, "interiorClean"=$11,
			"seatsCondition"=$12, "interiorDirt"=$13, smell=$14, "recordedById"=NULLIF($15,'')
			WHERE id = (SELECT id FROM "VehicleDailyRating" WHERE "vehicleId"=$1 AND "ratedDate"=$2::date ORDER BY "createdAt" LIMIT 1)`, args...)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "تعذر الحفظ — تأكد من السيارة")
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			if _, err := tx.Exec(`INSERT INTO "VehicleDailyRating" (id, "vehicleId", "ratedDate", wash, "exteriorClean", "exteriorCondition",
				"tireCondition", "glassClean", "lightsCondition", "technicalFaults", "faultDescription", "interiorClean",
				"seatsCondition", "interiorDirt", smell, "recordedById")
				VALUES (gen_random_uuid()::text, $1, $2::date, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NULLIF($15,''))`, args...); err != nil {
				WriteError(w, http.StatusBadRequest, "تعذر الحفظ — تأكد من السيارة")
				return
			}
		}
	}
	if err := tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
