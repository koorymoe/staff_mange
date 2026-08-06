package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type VehicleHandler struct {
	service *service.VehicleService
}

func NewVehicleHandler(s *service.VehicleService) *VehicleHandler {
	return &VehicleHandler{service: s}
}

func (h *VehicleHandler) List(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة المركبات")
		return
	}
	WriteJSON(w, http.StatusOK, vehicles)
}

// Options قائمة مختصرة للاختيار وقت تخصيص سيارة لحجز.
//
// شاشة تنسيق الحجوزات جانت تنادي /api/vehicles وهي محمية بصلاحية إدارة
// المركبات — وما عدها ولا موظف تنسيق. فالنتيجة: القائمة المنسدلة تطلع
// فاضية وما يكدرون يحددون سيارة أبداً.
//
// اختيار السيارة مو إدارة أسطول: نرجّع الاسم والرقم واللون بس، بلا
// وثائق ولا صيانة ولا مصاريف — فما نكشف شي بمجرد التخصيص.
func (h *VehicleHandler) Options(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة المركبات")
		return
	}
	type option struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		PlateNumber string `json:"plateNumber"`
		Color       *string `json:"color"`
	}
	out := []option{}
	for _, v := range vehicles {
		if !v.IsActive {
			continue
		}
		out = append(out, option{ID: v.ID, Name: v.Name, PlateNumber: v.PlateNumber, Color: v.Color})
	}
	WriteJSON(w, http.StatusOK, out)
}

func (h *VehicleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	vehicle, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, vehicle)
}

func (h *VehicleHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateVehicleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	vehicle, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, vehicle)
}

func (h *VehicleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VehicleHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := h.service.ListDocuments(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب وثائق السيارة")
		return
	}
	WriteJSON(w, http.StatusOK, docs)
}

func (h *VehicleHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleDocumentRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	doc, err := h.service.CreateDocument(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, doc)
}

func (h *VehicleHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteDocument(r.PathValue("id"), r.PathValue("docId")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حذف الوثيقة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *VehicleHandler) ListPhotos(w http.ResponseWriter, r *http.Request) {
	photos, err := h.service.ListPhotos(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب صور السيارة")
		return
	}
	WriteJSON(w, http.StatusOK, photos)
}

func (h *VehicleHandler) CreatePhoto(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehiclePhotoRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	photo, err := h.service.CreatePhoto(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, photo)
}

func (h *VehicleHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeletePhoto(r.PathValue("id"), r.PathValue("photoId")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حذف الصورة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *VehicleHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.ListLogs(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجلات السيارة")
		return
	}
	WriteJSON(w, http.StatusOK, logs)
}

func (h *VehicleHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleLogRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	log, err := h.service.CreateLog(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, log)
}

func (h *VehicleHandler) UpdateLog(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateVehicleLogRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	log, err := h.service.UpdateLog(r.PathValue("logId"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, log)
}

func (h *VehicleHandler) DeleteLog(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteLog(r.PathValue("logId")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حذف السجل")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// LogReceiptPhoto يرجع صورة الوصل لوحدها — منفصلة عن القائمة حتى ما تثقّلها.
func (h *VehicleHandler) LogReceiptPhoto(w http.ResponseWriter, r *http.Request) {
	photo, err := h.service.GetLogReceiptPhoto(r.PathValue("logId"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "السجل غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]*string{"receiptPhotoBase64": photo})
}

func (h *VehicleHandler) EmployeeFuelStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.EmployeeFuelStats(r.URL.Query().Get("vehicleId"), r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب إحصائية التعبئة")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

func (h *VehicleHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, err := h.service.ListIncidents(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أعطال/أضرار السيارة")
		return
	}
	WriteJSON(w, http.StatusOK, incidents)
}

func (h *VehicleHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleIncidentRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	incident, err := h.service.CreateIncident(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, incident)
}

func (h *VehicleHandler) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateVehicleIncidentRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	incident, err := h.service.UpdateIncident(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, incident)
}

func (h *VehicleHandler) ListMonthlyStatus(w http.ResponseWriter, r *http.Request) {
	statuses, err := h.service.ListMonthlyStatus(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حالة السيارة الشهرية")
		return
	}
	WriteJSON(w, http.StatusOK, statuses)
}

func (h *VehicleHandler) SetMonthlyStatus(w http.ResponseWriter, r *http.Request) {
	var req model.SetVehicleMonthlyStatusRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	status, err := h.service.SetMonthlyStatus(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, status)
}

// ── VehicleDailyRating (تقييم يومي + جودة غسيل الفنيين) ──

func (h *VehicleHandler) CreateDailyRating(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleDailyRatingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	rating, err := h.service.CreateDailyRating(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, rating)
}

func (h *VehicleHandler) ListDailyRatings(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	if since == "" {
		since = "1900-01-01"
	}
	ratings, err := h.service.ListDailyRatings(r.PathValue("id"), since)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تقييمات السيارة")
		return
	}
	WriteJSON(w, http.StatusOK, ratings)
}

func (h *VehicleHandler) VehicleScoreSummaries(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	if since == "" {
		since = "1900-01-01"
	}
	summaries, err := h.service.VehicleScoreSummaries(since)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ملخص تقييم السيارات")
		return
	}
	WriteJSON(w, http.StatusOK, summaries)
}

func (h *VehicleHandler) TechnicianWashSummaries(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	if since == "" {
		since = "1900-01-01"
	}
	summaries, err := h.service.TechnicianWashSummaries(since)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ملخص تقييم الفنيين")
		return
	}
	WriteJSON(w, http.StatusOK, summaries)
}

// ── VehicleIncidentAttachment ──

func (h *VehicleHandler) ListIncidentAttachments(w http.ResponseWriter, r *http.Request) {
	attachments, err := h.service.ListIncidentAttachments(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب مرفقات العطل/الضرر")
		return
	}
	WriteJSON(w, http.StatusOK, attachments)
}

func (h *VehicleHandler) CreateIncidentAttachment(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleIncidentAttachmentRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	attachment, err := h.service.CreateIncidentAttachment(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, attachment)
}

func (h *VehicleHandler) DeleteIncidentAttachment(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteIncidentAttachment(r.PathValue("id"), r.PathValue("attachmentId")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حذف المرفق")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── VehiclePart (إطارات وبطاريات) ──

func (h *VehicleHandler) ListParts(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.PathValue("id")
	vehicle, err := h.service.Get(vehicleID)
	currentOdometer := 0
	if err == nil && vehicle != nil {
		currentOdometer = vehicle.CurrentOdometer
	}
	parts, err := h.service.ListParts(vehicleID, currentOdometer)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قطع السيارة")
		return
	}
	WriteJSON(w, http.StatusOK, parts)
}

func (h *VehicleHandler) CreatePart(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehiclePartRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	part, err := h.service.CreatePart(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, part)
}

func (h *VehicleHandler) ReplacePart(w http.ResponseWriter, r *http.Request) {
	part, err := h.service.MarkPartReplaced(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تحديث القطعة")
		return
	}
	WriteJSON(w, http.StatusOK, part)
}

// ── تنبيهات الصيانة والوثائق ──

func (h *VehicleHandler) Alerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := h.service.VehicleAlerts()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تنبيهات السيارات")
		return
	}
	WriteJSON(w, http.StatusOK, alerts)
}

// ── ملخص مصاريف السيارة (المرحلة 3) ──

func (h *VehicleHandler) ExpenseSummary(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")
	summary, err := h.service.ExpenseSummary(r.PathValue("id"), month, year)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

// ── لوحة التحكم الشاملة للأسطول ──

func (h *VehicleHandler) FleetDashboard(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.FleetDashboard()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب لوحة تحكم الأسطول")
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}
