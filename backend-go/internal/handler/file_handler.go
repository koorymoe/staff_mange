package handler

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/storage"
)

// FileHandler رفع وعرض الملفات المخزّنة برّا قاعدة البيانات.
//
// المسار الي ينحفظ بقاعدة البيانات هو `/api/files/<key>` — يشتغل
// مباشرة داخل <img src> بالواجهة، بالضبط متل ما كانت تشتغل الـdata
// URLs. يعني الترحيل ما يكسر ولا شاشة.
type FileHandler struct {
	store  storage.Store
	secret []byte
	// employees: حالة صاحب الوسم (موقوف/محذوف = وسمه ما يفتح شي).
	employees *repository.EmployeeRepository
	// ⚠️ كاش دقيقة لحالة الموظف: الشاشة الوحدة تطلب عشرات الصور، واستعلام
	// لكل صورة يبطّئها. الإيقاف ياخذ أثره خلال دقيقة بالأكثر.
	mu    sync.Mutex
	cache map[string]fileAuthEntry
}

type fileAuthEntry struct {
	role   string
	active bool
	at     time.Time
}

func (h *FileHandler) employeeState(id string) (string, bool) {
	h.mu.Lock()
	if e, ok := h.cache[id]; ok && time.Since(e.at) < time.Minute {
		h.mu.Unlock()
		return e.role, e.active
	}
	h.mu.Unlock()
	emp, err := h.employees.FindByID(id)
	if err != nil {
		return "", false // خطأ قاعدة ما ينحفظ بالكاش — المحاولة الجاية تعيد
	}
	e := fileAuthEntry{at: time.Now()}
	if emp != nil {
		e.role, e.active = emp.Role, emp.Status == "ACTIVE"
	}
	h.mu.Lock()
	if h.cache == nil || len(h.cache) > 5000 {
		h.cache = map[string]fileAuthEntry{}
	}
	h.cache[id] = e
	h.mu.Unlock()
	return e.role, e.active
}

func NewFileHandler(s storage.Store, secret []byte, employees *repository.EmployeeRepository) *FileHandler {
	return &FileHandler{store: s, secret: secret, employees: employees}
}

// GET /api/files/token — وسم قصير العمر تستخدمه الواجهة بروابط الصور.
//
// الواجهة تجيبه مرة وحدة وتضيفه لكل رابط ملف — ما تحتاج نداء لكل صورة.
func (h *FileHandler) Token(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"token": storage.NewFileToken(h.secret, middleware.EmployeeIDFromContext(r))})
}

// POST /api/files — رفع ملف (multipart، الحقل اسمه file).
//
// المجلد ينجي من ?folder= ويتحدد بقائمة بيضاء — بدونها المستخدم
// يكتب أي مسار يريده.
func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(storage.MaxFileBytes); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر قراءة الملف — يمكن أكبر من الحد المسموح")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ماكو ملف بالطلب")
		return
	}
	defer file.Close()

	data, err := storage.ReadLimited(file, storage.MaxFileBytes)
	if err != nil {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}

	// النوع ينحدد من محتوى الملف نفسه — الترويسة المرسلة تنزوّر بسهولة
	contentType := storage.SniffContentType(data)
	if !storage.AllowedContentTypes[contentType] {
		WriteError(w, http.StatusBadRequest, "نوع الملف مو مسموح — صور (JPG/PNG/WEBP) أو PDF")
		return
	}

	key := storage.NewKey(safeFolder(r.URL.Query().Get("folder")), contentType)
	if err := h.store.Put(r.Context(), key, data, contentType); err != nil {
		log.Printf("file upload: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الملف")
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]any{
		"key":  key,
		"url":  "/api/files/" + key,
		"size": len(data),
		"type": contentType,
	})
}

// GET /api/files/{key...} — عرض الملف.
func (h *FileHandler) Serve(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/files/")
	if key == "" {
		WriteError(w, http.StatusBadRequest, "مفتاح الملف مطلوب")
		return
	}
	// الوسم بديل ترويسة Authorization لأن وسم <img> ما يرسلها
	employeeID, err := storage.VerifyFileToken(h.secret, r.URL.Query().Get("ft"))
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "وصول غير مصرّح للملف")
		return
	}
	// 🔴 (A-02) صاحب الوسم لازم يكون موظفاً فعّالاً هسه: الموقوف أو
	// المفصول ما يفتح ملف حتى لو بيده رابط بوسم ما انتهى.
	role := ""
	if employeeID != "" {
		empRole, active := h.employeeState(employeeID)
		if !active {
			WriteError(w, http.StatusUnauthorized, "وصول غير مصرّح للملف")
			return
		}
		role = empRole
	}
	// صور الكشف لها مسارها المحمي بالأدوار (/api/survey-photos/{id}/file)
	// — المسار العام ما يفتحها إلا للمدير، حتى ما ينلف عليها حارسها.
	if strings.HasPrefix(key, "surveys/") && role != "ADMIN" && role != "OWNER" {
		WriteError(w, http.StatusForbidden, "هذا الملف يفتح من شاشته بس")
		return
	}
	data, contentType, err := h.store.Get(r.Context(), key)
	if err != nil {
		if err == storage.ErrNotFound {
			WriteError(w, http.StatusNotFound, "الملف غير موجود")
			return
		}
		log.Printf("file serve %q: %v", key, err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الملف")
		return
	}
	w.Header().Set("Content-Type", contentType)
	// المفتاح عشوائي وما يتغيّر محتواه أبداً — فالتخزين المؤقت آمن وطويل
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	// ما نخلي المتصفح يخمّن النوع — يمنع تنفيذ ملف مرفوع كـHTML
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "inline")
	_, _ = w.Write(data)
}

// allowedFolders المجلدات المنطقية المعروفة. أي شي غيرها يروح misc.
var allowedFolders = map[string]bool{
	"products": true, "receipts": true, "vehicles": true, "projects": true,
	"reports": true, "exhibitions": true, "gps": true, "incidents": true,
	"misc": true, "sim": true,
}

func safeFolder(f string) string {
	f = strings.ToLower(strings.Trim(strings.TrimSpace(f), "/"))
	if allowedFolders[f] {
		return f
	}
	return "misc"
}
