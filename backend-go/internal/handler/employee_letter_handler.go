package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// EmployeeLetterHandler الطلبات — كتاب رسمي من الموظف للإدارة.
type EmployeeLetterHandler struct {
	repo   *repository.EmployeeLetterRepository
	notify *repository.NotificationRepository
}

func NewEmployeeLetterHandler(repo *repository.EmployeeLetterRepository, notify *repository.NotificationRepository) *EmployeeLetterHandler {
	return &EmployeeLetterHandler{repo: repo, notify: notify}
}

// GET /api/letters/addressees — الجهات الي ينوجّهلها الكتاب.
func (h *EmployeeLetterHandler) Addressees(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, model.LetterAddressees)
}

// POST /api/letters — الموظف يقدّم طلب.
func (h *EmployeeLetterHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateEmployeeLetterRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	if TextLen(req.Subject) < 3 {
		WriteError(w, http.StatusBadRequest, "اكتب موضوع الطلب")
		return
	}
	if TextLen(req.Body) < 10 {
		WriteError(w, http.StatusBadRequest, "اكتب تفاصيل الطلب — سطر واحد ما يوضّح شي للإدارة")
		return
	}
	if req.AddressedTo == "" {
		req.AddressedTo = model.LetterAddressees[0]
	}

	letter, err := h.repo.Create(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تقديم الطلب")
		return
	}
	// الإدارة تنبّه فوراً — الطلب الي يقعد بصندوق محد يفتحه نفس
	// الورقة الي تضيع على المكتب.
	if h.notify != nil {
		name := ""
		if letter.Employee != nil {
			name = letter.Employee.Name
		}
		msg := "📄 طلب جديد من " + name + ": " + req.Subject
		_ = h.notify.CreateForRole("ADMIN", "employee_letter", msg)
		_ = h.notify.CreateForRole("OWNER", "employee_letter", msg)
	}
	WriteJSON(w, http.StatusOK, letter)
}

// GET /api/letters/mine — طلبات الموظف نفسه.
func (h *EmployeeLetterHandler) Mine(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Mine(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلباتك")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/letters — صندوق الإدارة (للمالك ومدير النظام).
func (h *EmployeeLetterHandler) Inbox(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Inbox(r.URL.Query().Get("status"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الطلبات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/letters/{id}/decide — جواب الإدارة.
func (h *EmployeeLetterHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var req model.DecideEmployeeLetterRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	// الرفض بلا سبب ما يعلّم الموظف شي، ويخلي نفس الطلب يتكرر.
	if !req.Approve && TextLen(req.Note) < 5 {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الرفض — الموظف لازم يعرف ليش")
		return
	}

	letter, err := h.repo.Decide(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.notify != nil {
		verdict := "❌ انرفض طلبك: " + letter.Subject
		if req.Approve {
			verdict = "✅ انوافق على طلبك: " + letter.Subject
		}
		if req.Note != "" {
			verdict += " — " + req.Note
		}
		_ = h.notify.Create(letter.EmployeeID, "employee_letter_decision", verdict)
	}
	WriteJSON(w, http.StatusOK, letter)
}

// GET /api/letters/pending-count — شارة العدد بالقائمة.
func (h *EmployeeLetterHandler) PendingCount(w http.ResponseWriter, r *http.Request) {
	n, err := h.repo.PendingCount()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر العد")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"count": n})
}
