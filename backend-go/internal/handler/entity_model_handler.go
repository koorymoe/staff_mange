package handler

import (
	"log"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/storage"
)

// EntityModelHandler رفع وإدارة مجسّمات الكيان ثلاثية الأبعاد.
//
// 🔒 **الرفع بمسار خاص مو على `POST /api/files` العام.**
// المسار العام حارسه `requireAuth` — يعني **أي موظف** يگدر يرفع، وملف
// المجسّم يوصل ١٠ م.ب. فلو استخدمناه، أي حساب موظف يحشر تخزيننا
// بملفات ضخمة. وهنا الحارس **المالك حصراً**.
type EntityModelHandler struct {
	repo  *repository.EntityModelRepository
	store storage.Store
}

func NewEntityModelHandler(r *repository.EntityModelRepository, s storage.Store) *EntityModelHandler {
	return &EntityModelHandler{repo: r, store: s}
}

// GET /api/entity/models — القائمة للعرض والتبديل.
func (h *EntityModelHandler) List(w http.ResponseWriter, _ *http.Request) {
	rows, err := h.repo.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المجسّمات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/entity/models/active — شخصية النظام الحالية.
//
// ⚠️ **مسار لحاله مو ترشيح من القائمة**: هذا النداء يصير لـ**كل موظف**
// بكل تحميل للودجة، فلازم يرجّع صفاً واحداً مو القائمة كلها. ولأن
// الودجة تنبني قبل ما يختار (ع) شخصية، **ماكو نشط حالة طبيعية** ونرجّع
// `204` — والواجهة تقع على المجسّم المدمج بهدوء.
//
// والقراءة لأي موظف مسجّل (مو للمالك): كل موظف يحتاجها حتى يشوف شخصيته.
func (h *EntityModelHandler) Active(w http.ResponseWriter, _ *http.Request) {
	row, err := h.repo.GetActive()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب شخصية النظام")
		return
	}
	if row == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	WriteJSON(w, http.StatusOK, row)
}

// PUT /api/entity/models/{id}/activate — خلّي هذا المجسّم شخصية النظام.
//
// 🔒 **للمالك حصراً**: هذا يبدّل الشخصية الي يشوفها **كل** موظف بالنظام.
// و`id` = `builtin` يعني رجوعاً للمجسّم المدمج — يعني خطة تراجع بضغطة
// زر، بلا نشر ولا تعديل كود.
func (h *EntityModelHandler) Activate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "builtin" {
		id = ""
	}
	if err := h.repo.Activate(id); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/entity/models — رفع مجسّم (multipart: file + label).
//
// الرفع والتسجيل **بنداء واحد**: لو فصلناهم يصير ملف مخزون بلا صف
// يشير إله لو فشل النداء الثاني، ويبقى يشغّل مساحة بلا ما يبيّن بأي
// شاشة.
func (h *EntityModelHandler) Create(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	if err := r.ParseMultipartForm(storage.MaxFileBytes); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر قراءة الملف — يمكن أكبر من الحد المسموح (١٠ ميغا)")
		return
	}
	file, header, err := r.FormFile("file")
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

	// ⚠️ النوع من **محتوى الملف** مو من ترويسة المتصفح: الترويسة
	// تنزوّر بسهولة. وملف `.glb` بصمته الحرفان "glTF".
	contentType := storage.SniffContentType(data)
	if contentType != "model/gltf-binary" {
		WriteError(w, http.StatusBadRequest, "الملف لازم يكون مجسّماً بصيغة GLB")
		return
	}

	label := strings.TrimSpace(r.FormValue("label"))
	if label == "" {
		// اسم الملف بديل معقول — بس بلا الامتداد ولا أي مسار.
		label = strings.TrimSuffix(sanitizeName(header.Filename), ".glb")
	}
	if label == "" {
		label = "مجسّم"
	}

	key := storage.NewKey("models", contentType)
	if err := h.store.Put(r.Context(), key, data, contentType); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تخزين الملف")
		return
	}

	row, err := h.repo.Create(label, key, contentType, int64(len(data)), employeeID)
	if err != nil {
		// الصف فشل والملف مخزون — نحذفه حتى ما يبقى ملف يتيم.
		_ = h.store.Delete(r.Context(), key)
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// 🔴 **أول مجسّم يُرفَع يصير شخصية النظام تلقائياً.**
	//
	// السبب من الواقع: مالك النظام رفع/سحب وباق يشوف القديمة، لأن
	// الرفع والتفعيل چانوا **زرّين**. وهو ما يرفع مجسّماً حتى «يكون
	// موجوداً» — يرفعه حتى **يُعرض**. فالخطوة الثانية كانت مصيدة.
	//
	// ⚠️ **وبشرط ماكو نشط**: بلا الشرط، أي رفعة لاحقة — ولو تجريبية —
	// تبدّل الشخصية على شاشة **كل موظف** بلا ما يطلبها أحد. والتبديل
	// بعد أول واحد يبقى بالزر، وهو قرار واعٍ.
	if active, aerr := h.repo.GetActive(); aerr == nil && active == nil {
		// ⚠️ وفشل التفعيل **ما يفشّل الرفعة**: الملف مخزون والصف
		// موجود فعلاً، فإرجاع خطأ هنا يخلي المالك يعيد الرفع ويصير
		// عندنا صفّان لنفس المجسّم. نسجّلها ونكمل — والزر موجود.
		if err := h.repo.Activate(row.ID); err != nil {
			log.Printf("entity model auto-activate %q: %v", row.ID, err)
		} else {
			row.IsActive = true
		}
	}
	WriteJSON(w, http.StatusCreated, row)
}

// PUT /api/entity/models/{id}/archive — أرشفة ناعمة.
func (h *EntityModelHandler) Archive(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Archive(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// sanitizeName يشيل المسارات من اسم ملف جاي من المستخدم.
//
// ⚠️ أسماء الملفات من المتصفح ممكن تحمل `../` أو مسار ويندوز كامل —
// ونحنا نستخدمها **للعرض بس**، بس التنظيف واجب حتى لا تطلع بواجهة
// كأنها مسار داخلي.
func sanitizeName(n string) string {
	n = strings.ReplaceAll(n, "\\", "/")
	if i := strings.LastIndex(n, "/"); i >= 0 {
		n = n[i+1:]
	}
	return strings.TrimSpace(n)
}
