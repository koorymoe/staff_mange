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
	svc     *service.MatrixChainService
	resolve func(employeeID string)
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
