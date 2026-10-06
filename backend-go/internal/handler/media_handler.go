package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// ═══ الإعلام والعلاقات العامة — طلب (ع) 10-05 ═══
// المشروع يوصل مرحلة «📸 الإعلام» قبل التنفيذ ← تحويل للإعلام بكل التفاصيل
// (المهندس المشرف، الموقع، الموعد، المدة، متى يخلص). موظفي الإعلام يشوفون
// التحويلات، ويحددون موعد التصوير، ويأشّرون «صوّرنا» و«نشرنا» ويا الرابط.
type MediaHandler struct {
	repo    *repository.MediaRepository
	notify  func(msg string)
	resolve func(string)
}

func NewMediaHandler(repo *repository.MediaRepository, notify func(string)) *MediaHandler {
	return &MediaHandler{repo: repo, notify: notify}
}

const mediaStage = "📸 الإعلام"

func parseTime(s *string) *time.Time {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	for _, f := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02"} {
		if t, err := time.ParseInLocation(f, strings.TrimSpace(*s), time.FixedZone("Baghdad", 3*3600)); err == nil {
			return &t
		}
	}
	return nil
}

func trimPtr(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

type mediaTransferInput struct {
	EngineerID    *string `json:"engineerId"`
	StartAt       *string `json:"startAt"`
	ExpectedEndAt *string `json:"expectedEndAt"`
	Duration      *string `json:"duration"`
	Notes         *string `json:"notes"`
}

// POST /api/projects/{id}/media — تحويل المشروع للإعلام (ونقله لمرحلة الإعلام).
func (h *MediaHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	var in mediaTransferInput
	_ = json.NewDecoder(r.Body).Decode(&in)
	eng := trimPtr(in.EngineerID)
	if eng == nil {
		eng = h.repo.ProjectEngineer(projectID)
	}
	c := repository.MediaCreate{ProjectID: projectID, CreatedByID: middleware.EmployeeIDFromContext(r),
		EngineerID: eng, StartAt: parseTime(in.StartAt), ExpectedEndAt: parseTime(in.ExpectedEndAt),
		Duration: trimPtr(in.Duration), Notes: trimPtr(in.Notes)}
	// زر «📸 إعلام» مرة ثانية: يحدّث التحويل المفتوح بدل ما يكرره.
	if existing := h.repo.OpenForProject(projectID); existing != "" {
		if err := h.repo.UpdateBrief(existing, c); err != nil {
			WriteError(w, http.StatusBadRequest, "تعذر تحديث التحويل")
			return
		}
		WriteJSON(w, http.StatusOK, map[string]any{"id": existing, "updated": true})
		return
	}
	id, err := h.repo.Create(c)
	if err != nil {
		log.Printf("media transfer: %v", err)
		WriteError(w, http.StatusBadRequest, "تعذر التحويل للإعلام — تأكد إن المشروع موجود")
		return
	}
	// قرار (ع) 10-06: الإعلام زر بخطوات المشروع، مو مرحلة توقفه. المرحلة
	// تصير «📸 الإعلام» بس إذا المشروع بالعقد؛ إذا بالتنفيذ يبقى بمرحلته.
	if strings.Contains(h.repo.ProjectStage(projectID), "عقد") {
		_ = h.repo.SetProjectStage(projectID, mediaStage)
	}
	if b, err := h.repo.Get(id); err == nil && h.notify != nil {
		when := ""
		if b.StartAt != nil {
			when = " · يبدي " + b.StartAt.In(time.FixedZone("Baghdad", 3*3600)).Format("2006-01-02 15:04")
		}
		go h.notify(fmt.Sprintf("📸 مشروع جديد محوّل للإعلام: %s (%s)%s. التفاصيل بشاشة الإعلام.", b.ProjectName, b.ProjectCode, when))
	}
	WriteJSON(w, http.StatusCreated, map[string]any{"id": id})
}

// GET /api/media/briefs
func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.List()
	if err != nil {
		log.Printf("media list: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تحويلات الإعلام")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

type mediaUpdateInput struct {
	Status       string  `json:"status"`
	ShootAt      *string `json:"shootAt"`
	PublishedURL *string `json:"publishedUrl"`
	MediaNotes   *string `json:"mediaNotes"`
}

var mediaStatuses = map[string]bool{"NEW": true, "SCHEDULED": true, "SHOT": true, "PUBLISHED": true, "CANCELLED": true}

// PUT /api/media/briefs/{id} — موظف الإعلام يحدّث: موعد التصوير، صوّرنا، نشرنا.
func (h *MediaHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in mediaUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || !mediaStatuses[in.Status] {
		WriteError(w, http.StatusBadRequest, "حالة غير صحيحة")
		return
	}
	shoot := parseTime(in.ShootAt)
	if in.Status == "SCHEDULED" && shoot == nil {
		WriteError(w, http.StatusBadRequest, "حدد موعد التصوير")
		return
	}
	url := trimPtr(in.PublishedURL)
	if in.Status == "PUBLISHED" && url == nil {
		WriteError(w, http.StatusBadRequest, "حط رابط المنشور")
		return
	}
	me := middleware.EmployeeIDFromContext(r)
	if err := h.repo.Update(r.PathValue("id"), repository.MediaUpdate{Status: in.Status, ShootAt: shoot, MediaEmpID: &me,
		PublishedURL: url, MediaNotes: trimPtr(in.MediaNotes)}); err != nil {
		log.Printf("media update: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	b, _ := h.repo.Get(r.PathValue("id"))
	WriteJSON(w, http.StatusOK, b)
}

// ═══ سجل الانضباط الوظيفي (المدير والمالك) ═══
type DisciplineRecordHandler struct {
	svc *service.DisciplineRecordService
}

func NewDisciplineRecordHandler(svc *service.DisciplineRecordService) *DisciplineRecordHandler {
	return &DisciplineRecordHandler{svc: svc}
}

// GET /api/discipline-record?employeeId=
func (h *DisciplineRecordHandler) Get(w http.ResponseWriter, r *http.Request) {
	rec, err := h.svc.Build(r.URL.Query().Get("employeeId"))
	if err != nil {
		log.Printf("discipline record: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل الانضباط")
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}

// GET /api/projects/media/open — المشاريع المحوّلة للإعلام (علامة ✓ على الزر).
func (h *MediaHandler) OpenProjects(w http.ResponseWriter, r *http.Request) {
	ids, err := h.repo.OpenProjects()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, ids)
}

// GET /api/projects/{id}/media — التحويل المفتوح لهالمشروع (لتعبئة النافذة).
func (h *MediaHandler) ForProject(w http.ResponseWriter, r *http.Request) {
	id := h.repo.OpenForProject(r.PathValue("id"))
	if id == "" {
		WriteJSON(w, http.StatusOK, nil)
		return
	}
	b, err := h.repo.Get(id)
	if err != nil {
		WriteJSON(w, http.StatusOK, nil)
		return
	}
	WriteJSON(w, http.StatusOK, b)
}
