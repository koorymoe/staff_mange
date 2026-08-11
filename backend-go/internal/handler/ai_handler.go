package handler

import (
	"net/http"
	"strconv"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// AiHandler مسارات نواة الذكاء الاصطناعي.
//
// ⚠️ كل هذي المسارات محصورة بالمالك ومدير النظام (تنلف بـrequireAdmin
// من main). طلب صريح من صاحب العمل: «التقرير النهائي يطلع فقط للمدير
// النظام والمالك».
//
// وهذا مو تفصيل إداري — تحليل «ليش هذا الموظف وقّف الشغل» بيد زميله
// يتحول لسلاح داخلي، ويخلي الموظفين يخافون يكتبون السبب الحقيقي.
type AiHandler struct {
	repo  *repository.AiRepository
	brain *service.AiBrainService
}

func NewAiHandler(repo *repository.AiRepository, brain *service.AiBrainService) *AiHandler {
	return &AiHandler{repo: repo, brain: brain}
}

// GET /api/ai/signals?kind=&limit=
func (h *AiHandler) ListSignals(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := h.repo.ListSignals(r.URL.Query().Get("kind"), limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإشارات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/ai/process — يمشي بالإشارات المعلّقة يدوياً.
//
// يدوي بهاي المرحلة قصداً: قبل ما ننشترك بمنصّة، تشغيله بضغطة يخلي
// المالك يشوف النتيجة وقت ما يريد بدل ما يشتغل بالخلفية ويستهلك
// بلا ما أحد ينتبه.
func (h *AiHandler) Process(w http.ResponseWriter, r *http.Request) {
	n, err := h.brain.Process(50)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"analyzed": n})
}

// GET /api/ai/metrics?from=&to=
func (h *AiHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	parse := func(v string, def time.Time) time.Time {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return t
		}
		return def
	}
	to := parse(r.URL.Query().Get("to"), time.Now())
	from := parse(r.URL.Query().Get("from"), to.AddDate(0, -1, 0))
	rows, err := h.repo.ListMetrics(from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المؤشرات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/ai/work-window · PUT /api/ai/work-window
//
// ساعات الدوام تغذّي التحليل: «وقّف الساعة ١١:٥٠ ليلاً» تفسير مختلف
// تماماً عن «وقّف الساعة ١١ صباحاً».
func (h *AiHandler) GetWorkWindow(w http.ResponseWriter, r *http.Request) {
	win, err := h.repo.WorkWindow()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ساعات الدوام")
		return
	}
	WriteJSON(w, http.StatusOK, win)
}

func (h *AiHandler) SetWorkWindow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StartHour int `json:"startHour"`
		EndHour   int `json:"endHour"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	// ⚠️ ١٢ ليلاً = ٢٤ مو ٠: الصفر يخلي «نهاية الدوام» قبل بدايته وكل
	// حساب «باقي شكد على النهاية» يطلع بالسالب.
	if req.StartHour < 0 || req.StartHour > 23 || req.EndHour < 1 || req.EndHour > 24 {
		WriteError(w, http.StatusBadRequest, "البداية ٠-٢٣ والنهاية ١-٢٤ (١٢ ليلاً = ٢٤)")
		return
	}
	if req.EndHour <= req.StartHour {
		WriteError(w, http.StatusBadRequest, "نهاية الدوام لازم تكون بعد بدايته")
		return
	}
	if err := h.repo.SetWorkWindow(req.StartHour, req.EndHour); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/ai/catalog — شنو يعرف النظام يحلله، وشنو لسه ينتظر المنصّة.
//
// شاشة الهيكلة تقراه بدل ما تعيد كتابة نفس القوائم بالواجهة — وإلا
// أي إشارة جديدة بالسيرفر ما تظهر لحد ما أحد يتذكر يحدّث الواجهة.
func (h *AiHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	type item struct {
		Key    string `json:"key"`
		Label  string `json:"label"`
		Ready  bool   `json:"ready"`
		Detail string `json:"detail"`
	}
	signals := []item{
		{model.AiSignalWorkStopped, model.AiSignalLabel(model.AiSignalWorkStopped), true,
			"يجمع الأدلة كاملة: الساعة ونهاية الدوام، طلبات المواد، سلة الزبون، وسجل الموظف"},
		{model.AiSignalLateStart, model.AiSignalLabel(model.AiSignalLateStart), false,
			"ينتظر: ربط وقت الخروج بالموعد"},
		{model.AiSignalRepeatPostpone, model.AiSignalLabel(model.AiSignalRepeatPostpone), false,
			"ينتظر: جامع أدلة التأجيل"},
		{model.AiSignalInvoiceAdjusted, model.AiSignalLabel(model.AiSignalInvoiceAdjusted), false,
			"ينتظر: جامع أدلة الفواتير"},
		{model.AiSignalRepeatPartial, model.AiSignalLabel(model.AiSignalRepeatPartial), false,
			"ينتظر: جامع أدلة الإنجاز الجزئي"},
	}
	metrics := []item{
		{model.AiMetricStopRate, model.AiMetricLabel(model.AiMetricStopRate), false, "ينتظر: حاسبة المؤشرات"},
		{model.AiMetricStopMinutesAvg, model.AiMetricLabel(model.AiMetricStopMinutesAvg), false, "ينتظر: حاسبة المؤشرات"},
		{model.AiMetricMaterialMissRate, model.AiMetricLabel(model.AiMetricMaterialMissRate), false, "ينتظر: حاسبة المؤشرات"},
		{model.AiMetricScopeCreepRate, model.AiMetricLabel(model.AiMetricScopeCreepRate), false, "ينتظر: حاسبة المؤشرات"},
		{model.AiMetricProcurementDelay, model.AiMetricLabel(model.AiMetricProcurementDelay), false, "ينتظر: حاسبة المؤشرات"},
		{model.AiMetricLateStartRate, model.AiMetricLabel(model.AiMetricLateStartRate), false, "ينتظر: حاسبة المؤشرات"},
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"signals": signals,
		"metrics": metrics,
		// الحاكم الحالي: قواعد حتمية. تنبدل بالمنصّة بسطر واحد.
		"judge":          "rules-v1",
		"platformLinked": false,
	})
}
