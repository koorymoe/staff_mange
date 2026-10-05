package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ صوت الموظفين + المشاكل الوظيفية ═══
// الموظف: فضفضته هو بس. المدير/المالك: كل النصوص. المراقب: تقارير التعمّق
// والمشاكل الوظيفية. **والزميل المقصود ما عنده أي مسار يشوف بيه شي.**
type PeerVoiceHandler struct{ svc *service.PeerVoiceService }

func NewPeerVoiceHandler(svc *service.PeerVoiceService) *PeerVoiceHandler {
	return &PeerVoiceHandler{svc: svc}
}

func isAdminRole(r *http.Request) bool {
	role := middleware.RoleFromContext(r)
	return role == "ADMIN" || role == "OWNER"
}

// GET /api/peer/me
func (h *PeerVoiceHandler) Mine(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.Mine(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, st)
}

// POST /api/peer/me
func (h *PeerVoiceHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var in service.PeerCheckinIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	st, err := h.svc.Submit(middleware.EmployeeIDFromContext(r), in)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, st)
}

// POST /api/peer/notes/{id}/followup — أجوبة أسئلة ماتركس.
func (h *PeerVoiceHandler) Followup(w http.ResponseWriter, r *http.Request) {
	var in struct{ What, Why, Wish string }
	_ = json.NewDecoder(r.Body).Decode(&in)
	n, err := h.svc.Followup(middleware.EmployeeIDFromContext(r), r.PathValue("id"), in.What, in.Why, in.Wish)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, n)
}

// POST /api/peer/notes/{id}/accept
func (h *PeerVoiceHandler) Accept(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Accepted bool   `json:"accepted"`
		Note     string `json:"note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	n, err := h.svc.Accept(middleware.EmployeeIDFromContext(r), r.PathValue("id"), in.Accepted, in.Note)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, n)
}

// GET /api/peer/voice — المدير/المالك كامل، والمراقب تقارير التعمّق.
func (h *PeerVoiceHandler) Report(w http.ResponseWriter, r *http.Request) {
	rep, err := h.svc.Report(isAdminRole(r))
	if err != nil {
		log.Printf("peer voice: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب صوت الموظفين")
		return
	}
	WriteJSON(w, http.StatusOK, rep)
}

// ── المشاكل الوظيفية (المراقب والمدير والمالك) ──

func (h *PeerVoiceHandler) Issues(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.Issues()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func (h *PeerVoiceHandler) Issue(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.Issue(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, d)
}

func (h *PeerVoiceHandler) OpenIssue(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title        string  `json:"title"`
		PartyAID     string  `json:"partyAId"`
		PartyBID     *string `json:"partyBId"`
		Description  string  `json:"description"`
		Severity     string  `json:"severity"`
		SourceNoteID *string `json:"sourceNoteId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if in.PartyBID != nil && *in.PartyBID == "" {
		in.PartyBID = nil
	}
	id, err := h.svc.OpenIssue(middleware.EmployeeIDFromContext(r), in.Title, in.PartyAID, in.PartyBID, in.Description, in.Severity, in.SourceNoteID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *PeerVoiceHandler) AddEntry(w http.ResponseWriter, r *http.Request) {
	var in struct{ Kind, Text string }
	_ = json.NewDecoder(r.Body).Decode(&in)
	if err := h.svc.AddEntry(r.PathValue("id"), middleware.EmployeeIDFromContext(r), in.Kind, in.Text); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	d, _ := h.svc.Issue(r.PathValue("id"))
	WriteJSON(w, http.StatusOK, d)
}

func (h *PeerVoiceHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Analyze(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	d, _ := h.svc.Issue(r.PathValue("id"))
	WriteJSON(w, http.StatusOK, d)
}

func (h *PeerVoiceHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Decision   string  `json:"decision"`
		Status     string  `json:"status"`
		FollowUpAt *string `json:"followUpAt"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	var fu *time.Time
	if in.FollowUpAt != nil && *in.FollowUpAt != "" {
		if t, err := time.ParseInLocation("2006-01-02", *in.FollowUpAt, time.FixedZone("Baghdad", 3*3600)); err == nil {
			fu = &t
		}
	}
	if err := h.svc.Decide(r.PathValue("id"), middleware.EmployeeIDFromContext(r), in.Decision, in.Status, fu); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	d, _ := h.svc.Issue(r.PathValue("id"))
	WriteJSON(w, http.StatusOK, d)
}

// POST /api/workplace-issues/{id}/party-followup — الطرف نفسه يجاوب المتابعة.
func (h *PeerVoiceHandler) PartyFollowup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Better bool   `json:"better"`
		Text   string `json:"text"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if err := h.svc.PartyFollowup(r.PathValue("id"), middleware.EmployeeIDFromContext(r), in.Better, in.Text); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}
