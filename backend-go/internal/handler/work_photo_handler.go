package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/storage"
)

// WorkPhotoHandler صور «قبل» و«بعد» لشغلة حجز (دازاي 10-25: كادر الديكور يصوّر
// بدل ما يكتب). المسارات وراء requireBookingParty — كادر الحجز ومن يديره.
type WorkPhotoHandler struct {
	db    *sqlx.DB
	store storage.Store
}

func NewWorkPhotoHandler(db *sqlx.DB, store storage.Store) *WorkPhotoHandler {
	return &WorkPhotoHandler{db: db, store: store}
}

type workPhoto struct {
	ID          string    `db:"id" json:"id"`
	Phase       string    `db:"phase" json:"phase"`
	EmployeeID  *string   `db:"employeeId" json:"employeeId"`
	FileKey     string    `db:"fileKey" json:"-"`
	ContentType string    `db:"contentType" json:"-"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
}

// GET /api/bookings/{id}/work-photos
func (h *WorkPhotoHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []workPhoto{}
	if err := h.db.Select(&rows, `SELECT id, phase, "employeeId", "fileKey", "contentType", "createdAt"
		FROM "WorkPhoto" WHERE "bookingId" = $1 ORDER BY "createdAt"`, r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الصور")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/bookings/{id}/work-photos — multipart: file + phase (BEFORE|AFTER)
func (h *WorkPhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(storage.MaxFileBytes); err != nil {
		WriteError(w, http.StatusBadRequest, "الصورة كبيرة هواية — جرّب صورة ثانية")
		return
	}
	phase := strings.ToUpper(strings.TrimSpace(r.FormValue("phase")))
	if phase != "BEFORE" && phase != "AFTER" {
		WriteError(w, http.StatusBadRequest, "لازم تحدد: صورة قبل لو بعد")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ماكو صورة")
		return
	}
	defer file.Close()
	data, err := storage.ReadLimited(file, storage.MaxFileBytes)
	if err != nil {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	ct := storage.SniffContentType(data)
	if !strings.HasPrefix(ct, "image/") {
		WriteError(w, http.StatusBadRequest, "صور بس")
		return
	}
	key := storage.NewKey("work-photos", ct)
	if err := h.store.Put(r.Context(), key, data, ct); err != nil {
		log.Printf("work photo upload: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الصورة")
		return
	}
	var p workPhoto
	if err := h.db.Get(&p, `INSERT INTO "WorkPhoto" ("bookingId", "employeeId", phase, "fileKey", "contentType")
		VALUES ($1, $2, $3, $4, $5) RETURNING id, phase, "employeeId", "fileKey", "contentType", "createdAt"`,
		r.PathValue("id"), middleware.EmployeeIDFromContext(r), phase, key, ct); err != nil {
		_ = h.store.Delete(r.Context(), key)
		WriteError(w, http.StatusBadRequest, "تعذر ربط الصورة بالحجز")
		return
	}
	WriteJSON(w, http.StatusCreated, p)
}

// GET /api/bookings/{id}/work-photos/{photoId}/file
func (h *WorkPhotoHandler) File(w http.ResponseWriter, r *http.Request) {
	var p workPhoto
	if err := h.db.Get(&p, `SELECT id, phase, "employeeId", "fileKey", "contentType", "createdAt"
		FROM "WorkPhoto" WHERE id = $1 AND "bookingId" = $2`, r.PathValue("photoId"), r.PathValue("id")); err != nil {
		WriteError(w, http.StatusNotFound, "الصورة مو موجودة")
		return
	}
	data, ct, err := h.store.Get(r.Context(), p.FileKey)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الصورة")
		return
	}
	if ct == "" {
		ct = p.ContentType
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write(data)
}
