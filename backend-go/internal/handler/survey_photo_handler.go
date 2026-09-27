package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/storage"
)

// SurveyPhotoHandler صور الكشف — للحجز والمشروع.
//
// ⚠️ الفحص كله بالهاندلر مو بميدل وير RequireRole: ذاك يسجّل مخالفة
// ويقفل الحساب بعد خمس، والليدر الي رفع الصور لازم يشوفها بلا ما
// يُحسب عليه تجاوز.
type SurveyPhotoHandler struct {
	repo  *repository.SurveyPhotoRepository
	perms *repository.PermissionRepository
	store storage.Store
}

func NewSurveyPhotoHandler(repo *repository.SurveyPhotoRepository, perms *repository.PermissionRepository, store storage.Store) *SurveyPhotoHandler {
	return &SurveyPhotoHandler{repo: repo, perms: perms, store: store}
}

// الي يشوفون وينزّلون صور أي كشف: المراقب، المحاسب/المدقق، المدير
// والمالك، وإداري الحجوزات.
var (
	surveyViewRoles = map[string]bool{"ADMIN": true, "OWNER": true, "MONITOR": true, "FINANCE": true, "HR_COORDINATOR": true}
	surveyViewPerms = []string{"monitoring", "auditing", "finance", "finance_audit", "coordinator", "crew_management"}
	// يرفعون على أي حجز كشف (غير الكادر نفسه): الإداري والمدير.
	surveyUploadRoles = map[string]bool{"ADMIN": true, "OWNER": true, "HR_COORDINATOR": true}
	surveyUploadPerms = []string{"coordinator", "crew_management"}
	projectMgrRoles   = map[string]bool{"PROJECT_MANAGER": true}
	projectMgrPerms   = []string{"project_management"}
)

func (h *SurveyPhotoHandler) hasAny(r *http.Request, roles map[string]bool, perms []string) bool {
	if roles[middleware.RoleFromContext(r)] {
		return true
	}
	for _, p := range perms {
		if ok, err := h.perms.HasPermission(middleware.EmployeeIDFromContext(r), p); err == nil && ok {
			return true
		}
	}
	return false
}

// isTeam: الكادر الي طلع للكشف (حجز) أو فريق كشف المشروع.
func (h *SurveyPhotoHandler) isTeam(r *http.Request, bookingID, projectID string) bool {
	me := middleware.EmployeeIDFromContext(r)
	var ok bool
	var err error
	if bookingID != "" {
		ok, err = h.repo.IsBookingCrew(bookingID, me)
	} else {
		ok, err = h.repo.IsProjectSurveyTeam(projectID, me)
	}
	return err == nil && ok
}

func (h *SurveyPhotoHandler) canView(r *http.Request, bookingID, projectID string) bool {
	if h.hasAny(r, surveyViewRoles, surveyViewPerms) {
		return true
	}
	if projectID != "" && h.hasAny(r, projectMgrRoles, projectMgrPerms) {
		return true
	}
	return h.isTeam(r, bookingID, projectID)
}

func (h *SurveyPhotoHandler) canUpload(r *http.Request, bookingID, projectID string) bool {
	if h.hasAny(r, surveyUploadRoles, surveyUploadPerms) {
		return true
	}
	if projectID != "" && h.hasAny(r, projectMgrRoles, projectMgrPerms) {
		return true
	}
	return h.isTeam(r, bookingID, projectID)
}

func ownerIDs(bookingID, projectID string) (*string, *string, error) {
	switch {
	case bookingID != "" && projectID == "":
		return &bookingID, nil, nil
	case projectID != "" && bookingID == "":
		return nil, &projectID, nil
	}
	return nil, nil, errors.New("لازم حجز أو مشروع واحد")
}

// GET /api/survey-photos?bookingId=|projectId=
func (h *SurveyPhotoHandler) List(w http.ResponseWriter, r *http.Request) {
	bookingID := r.URL.Query().Get("bookingId")
	projectID := r.URL.Query().Get("projectId")
	if _, _, err := ownerIDs(bookingID, projectID); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !h.canView(r, bookingID, projectID) {
		WriteError(w, http.StatusForbidden, "صور الكشف للمراقب والمحاسب والمدير وإداري الحجوزات")
		return
	}
	var photos []repository.SurveyPhoto
	var err error
	if bookingID != "" {
		photos, err = h.repo.ListForBooking(bookingID)
	} else {
		photos, err = h.repo.ListForProject(projectID)
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب صور الكشف")
		return
	}
	WriteJSON(w, http.StatusOK, photos)
}

// POST /api/survey-photos — multipart: file + bookingId أو projectId.
func (h *SurveyPhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(storage.MaxFileBytes); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر قراءة الصورة — يمكن أكبر من الحد المسموح")
		return
	}
	bookingID := strings.TrimSpace(r.FormValue("bookingId"))
	projectID := strings.TrimSpace(r.FormValue("projectId"))
	bID, pID, err := ownerIDs(bookingID, projectID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !h.canUpload(r, bookingID, projectID) {
		WriteError(w, http.StatusForbidden, "رفع صور الكشف للي طلع للكشف أو الإداري")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ماكو صورة بالطلب")
		return
	}
	defer file.Close()
	data, err := storage.ReadLimited(file, storage.MaxFileBytes)
	if err != nil {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	contentType := storage.SniffContentType(data)
	if !strings.HasPrefix(contentType, "image/") {
		WriteError(w, http.StatusBadRequest, "صور بس (JPG/PNG/WEBP)")
		return
	}
	key := storage.NewKey("surveys", contentType)
	if err := h.store.Put(r.Context(), key, data, contentType); err != nil {
		log.Printf("survey photo upload: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الصورة")
		return
	}
	photo, err := h.repo.Create(bID, pID, key, contentType, header.Filename, middleware.EmployeeIDFromContext(r))
	if err != nil {
		_ = h.store.Delete(r.Context(), key)
		WriteError(w, http.StatusBadRequest, "تعذر ربط الصورة — تأكد الحجز/المشروع موجود")
		return
	}
	WriteJSON(w, http.StatusCreated, photo)
}

func (h *SurveyPhotoHandler) load(w http.ResponseWriter, r *http.Request) *repository.SurveyPhoto {
	photo, err := h.repo.Get(r.PathValue("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteError(w, http.StatusNotFound, "الصورة غير موجودة")
		} else {
			WriteError(w, http.StatusInternalServerError, "تعذر جلب الصورة")
		}
		return nil
	}
	return photo
}

// GET /api/survey-photos/{id}/file[?download=1]
func (h *SurveyPhotoHandler) File(w http.ResponseWriter, r *http.Request) {
	photo := h.load(w, r)
	if photo == nil {
		return
	}
	var bookingID, projectID string
	if photo.BookingID != nil {
		bookingID = *photo.BookingID
	}
	if photo.ProjectID != nil {
		projectID = *photo.ProjectID
	}
	if !h.canView(r, bookingID, projectID) {
		WriteError(w, http.StatusForbidden, "صور الكشف للمراقب والمحاسب والمدير وإداري الحجوزات")
		return
	}
	data, contentType, err := h.store.Get(r.Context(), photo.FileKey)
	if err != nil {
		WriteError(w, http.StatusNotFound, "ملف الصورة غير موجود")
		return
	}
	disposition := "inline"
	if r.URL.Query().Get("download") == "1" {
		name := photo.FileName
		if name == "" {
			name = "survey-" + photo.ID
		}
		disposition = "attachment; filename*=UTF-8''" + url.PathEscape(name)
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write(data)
}

// DELETE /api/survey-photos/{id} — الي رفعها أو المدير/المالك.
func (h *SurveyPhotoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	photo := h.load(w, r)
	if photo == nil {
		return
	}
	role := middleware.RoleFromContext(r)
	if photo.UploadedByID != middleware.EmployeeIDFromContext(r) && role != "ADMIN" && role != "OWNER" {
		WriteError(w, http.StatusForbidden, "الحذف للي رفع الصورة أو المدير")
		return
	}
	if err := h.repo.Delete(photo.ID); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حذف الصورة")
		return
	}
	_ = h.store.Delete(r.Context(), photo.FileKey)
	w.WriteHeader(http.StatusNoContent)
}
