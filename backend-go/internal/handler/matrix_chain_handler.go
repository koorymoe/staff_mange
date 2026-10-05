package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ ماتركس ٢٠٥٠ — سلسلة الحجز وتقارير الأدوار وتقييم الفنيين ═══
type MatrixChainHandler struct {
	svc            *service.MatrixChainService
	resolve        func(employeeID string)
	notifyShortage func(msg string)
}

func NewMatrixChainHandler(svc *service.MatrixChainService, resolve func(string)) *MatrixChainHandler {
	return &MatrixChainHandler{svc: svc, resolve: resolve}
}

// GET /api/ai/chain/{bookingId} — المدير والمراقب.
func (h *MatrixChainHandler) Chain(w http.ResponseWriter, r *http.Request) {
	ch, err := h.svc.Build(r.PathValue("bookingId"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, ch)
}

// GET /api/ai/role-chain?role=&days= — المدير والمالك بس (فيه تقرير المراقبين).
func (h *MatrixChainHandler) RoleChain(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	rep, err := h.svc.RoleReport(strings.ToUpper(r.URL.Query().Get("role")), days)
	if err != nil {
		log.Printf("role chain: %v", err)
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rep)
}

// GET /api/ai/chain-learning — الحدود الي تعلّمها ماتركس.
func (h *MatrixChainHandler) Learning(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.svc.Learning())
}

// GET /api/ai/role-chain/roles — قائمة الأدوار بالترتيب.
func (h *MatrixChainHandler) Roles(w http.ResponseWriter, r *http.Request) {
	out := []map[string]string{}
	for _, k := range service.ChainRoleOrder {
		out = append(out, map[string]string{"key": k, "title": service.ChainRoleTitles[k]})
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/crew-ratings/pending — حجوزات الليدر الي تنتظر تقييمه.
func (h *MatrixChainHandler) Pending(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Repo().PendingRatings(middleware.EmployeeIDFromContext(r))
	if err != nil {
		log.Printf("pending ratings: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التقييمات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

type crewRatingInput struct {
	BookingID string `json:"bookingId"`
	Ratings   []struct {
		TechnicianID string  `json:"technicianId"`
		Score        int     `json:"score"`
		Note         *string `json:"note"`
	} `json:"ratings"`
}

// POST /api/crew-ratings — الليدر يقيّم فنيّي حجزه هو بس.
func (h *MatrixChainHandler) Rate(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	var in crewRatingInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.BookingID == "" || len(in.Ratings) == 0 {
		WriteError(w, http.StatusBadRequest, "بيانات التقييم ناقصة")
		return
	}
	repo := h.svc.Repo()
	if !repo.IsBookingLeader(in.BookingID, me) {
		WriteError(w, http.StatusForbidden, "تگدر تقيّم بس فنيّي حجز إنت ليدره")
		return
	}
	for _, x := range in.Ratings {
		if x.Score < 1 || x.Score > 5 || x.TechnicianID == me || !repo.IsBookingTech(in.BookingID, x.TechnicianID) {
			WriteError(w, http.StatusBadRequest, "تقييم غير صالح (١–٥ ولفني مكلّف بالحجز)")
			return
		}
	}
	for _, x := range in.Ratings {
		if x.Note != nil {
			t := strings.TrimSpace(*x.Note)
			x.Note = &t
		}
		if err := repo.SaveRating(in.BookingID, me, x.TechnicianID, x.Score, x.Note); err != nil {
			log.Printf("save rating: %v", err)
			WriteError(w, http.StatusInternalServerError, "تعذر حفظ التقييم")
			return
		}
	}
	if h.resolve != nil {
		go h.resolve(me)
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ═══ عين ماتركس على المشاريع ═══
type MatrixProjectHandler struct{ svc *service.MatrixProjectService }

func NewMatrixProjectHandler(svc *service.MatrixProjectService) *MatrixProjectHandler {
	return &MatrixProjectHandler{svc: svc}
}

// GET /api/ai/projects — كل المشاريع بحكم ماتركس + الأشخاص.
func (h *MatrixProjectHandler) Report(w http.ResponseWriter, r *http.Request) {
	rep, err := h.svc.Report()
	if err != nil {
		log.Printf("matrix projects: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر تحليل المشاريع")
		return
	}
	WriteJSON(w, http.StatusOK, rep)
}

// GET /api/ai/projects/{id} — ترتيب العمل بمشروع واحد.
func (h *MatrixProjectHandler) Chain(w http.ResponseWriter, r *http.Request) {
	ch, err := h.svc.Chain(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, ch)
}

// ═══ جرد العدّة بعد الحجز ═══

// SetShortageNotifier ينبّه صاحب صلاحية المخزن إذا الفني لگه نقص بعد الشغل.
func (h *MatrixChainHandler) SetShortageNotifier(f func(msg string)) { h.notifyShortage = f }

// GET /api/inventory/after-checks/pending — حجوزاتي المنجزة الي تنتظر جردي بعدها.
func (h *MatrixChainHandler) AfterPending(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Repo().AfterInventoryPending(middleware.EmployeeIDFromContext(r))
	if err != nil {
		log.Printf("after inventory pending: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الجرد")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

type afterInventoryInput struct {
	BookingID    string  `json:"bookingId"`
	Complete     bool    `json:"complete"`
	MissingItems *string `json:"missingItems"`
}

// POST /api/inventory/after-checks — الفني يجرد عدّته بعد حجز هو مكلّف بيه.
func (h *MatrixChainHandler) AfterSave(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	var in afterInventoryInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.BookingID == "" {
		WriteError(w, http.StatusBadRequest, "بيانات الجرد ناقصة")
		return
	}
	repo := h.svc.Repo()
	if !repo.IsBookingTech(in.BookingID, me) {
		WriteError(w, http.StatusForbidden, "تگدر تجرد بس بعد حجز إنت مكلّف بيه")
		return
	}
	var missing *string
	if in.MissingItems != nil {
		if t := strings.TrimSpace(*in.MissingItems); t != "" {
			missing = &t
		}
	}
	if !in.Complete && missing == nil {
		WriteError(w, http.StatusBadRequest, "اكتب شنو الناقص")
		return
	}
	if err := repo.SaveAfterInventory(in.BookingID, me, in.Complete, missing); err != nil {
		log.Printf("after inventory save: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الجرد")
		return
	}
	if !in.Complete && h.notifyShortage != nil {
		code, name := repo.BookingCodeAndEmployee(in.BookingID, me)
		go h.notifyShortage("🧰 نقص بالعدّة بعد الحجز " + code + " (" + name + "): " + *missing)
	}
	if h.resolve != nil {
		go h.resolve(me)
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ═══ ماتركس يقترح (المرحلة الثانية) ═══
type MatrixSuggestHandler struct{ svc *service.MatrixSuggestService }

func NewMatrixSuggestHandler(svc *service.MatrixSuggestService) *MatrixSuggestHandler {
	return &MatrixSuggestHandler{svc: svc}
}

// GET /api/ai/suggest/{bookingId} — اقتراح الموعد والكادر للمنسق.
func (h *MatrixSuggestHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.Suggest(r.PathValue("bookingId"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, s)
}

// GET /api/ai/suggest-accuracy?days= — دقة اقتراحات ماتركس. المدير والمالك.
func (h *MatrixSuggestHandler) Accuracy(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	a, err := h.svc.Accuracy(days)
	if err != nil {
		log.Printf("suggest accuracy: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر حساب الدقة")
		return
	}
	WriteJSON(w, http.StatusOK, a)
}

// ═══ المرحلة الثالثة: ماتركس ينفّذ ═══
type MatrixAutonomyHandler struct{ svc *service.MatrixAutonomyService }

func NewMatrixAutonomyHandler(svc *service.MatrixAutonomyService) *MatrixAutonomyHandler {
	return &MatrixAutonomyHandler{svc: svc}
}

// GET /api/ai/autonomy — المفاتيح وبوابة الدقة وآخر الأفعال. المدير والمالك.
func (h *MatrixAutonomyHandler) Status(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.svc.Status())
}

// GET /api/ai/autonomy/crew — تكليفات ماتركس الحية (شارة المنسق).
func (h *MatrixAutonomyHandler) ActiveCrew(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.svc.ActiveAutoCrew())
}

// POST /api/ai/autonomy/crew/{id}/undo — المنسق يتراجع عن تكليف ماتركس.
func (h *MatrixAutonomyHandler) UndoCrew(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.UndoAutoCrew(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}
