package handler

import (
	"encoding/json"
	"net/http"
	"staffmange-api/internal/errlog"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/middleware"
)

// ═══ 🗑️ المحذوفات — المالك يشوف أي سجل شغل انحذف ويرجّعه (تعديلات ٦ج — 11-04) ═══
// المصدر: سجل التدقيق (AuditTrail، action=DELETE) — الصف محفوظ كامل لحظة الحذف.

// RestorableEntities الجداول الي يرجع منها (القسم ١ + جداول الفلوس والصلاحيات).
var RestorableEntities = map[string]string{
	"Project": "مشروع", "Quotation": "عرض سعر", "QuotationItem": "بند عرض سعر", "Customer": "زبون",
	"RemoteWorkEntry": "ساعات من البيت", "ToolRequest": "طلب عُدّة", "SurveyPhoto": "صورة كشف", "KpiEvaluation": "تقييم KPI",
	"VehicleLog": "سجل مركبة", "VehicleWashLog": "غسل مركبة", "VehicleDocument": "وثيقة مركبة", "VehiclePhoto": "صورة مركبة",
	"VehicleIncidentAttachment": "مرفق حادث", "Vehicle": "مركبة", "ItAsset": "جهاز IT", "PersonalTool": "عُدّة شخصية",
	"VehicleTool": "عُدّة مركبة", "Product": "منتج", "Service": "خدمة", "Supplier": "مورد", "TechSupplier": "مورد تقني",
	"VipCustomer": "زبون VIP", "ProcurementQuote": "عرض شراء", "TrainingProgram": "برنامج تدريب",
	"TrainingProgramParticipant": "مشارك بتدريب", "Announcement": "إعلان",
	"LeaderInvoice": "فاتورة ليدر", "LeaderInvoiceMaterialItem": "مادة فاتورة", "LeaderInvoiceAdjustment": "تعديل فاتورة",
	"ProjectPayment": "دفعة مشروع", "RevolvingFundTxn": "حركة دوار", "CreditSale": "آجل", "CreditPayment": "دفعة آجل",
	"Expense": "مصروف", "EmployeePermission": "صلاحية موظف", "OrgAssignment": "منصب",
}

type DeletedRecordsHandler struct{ db *sqlx.DB }

func NewDeletedRecordsHandler(db *sqlx.DB) *DeletedRecordsHandler {
	return &DeletedRecordsHandler{db: db}
}

type deletedRow struct {
	ID        int64           `db:"id" json:"id"`
	At        time.Time       `db:"at" json:"at"`
	Entity    string          `db:"entity" json:"entity"`
	Label     string          `db:"-" json:"label"`
	EntityID  *string         `db:"entityId" json:"entityId"`
	ActorName *string         `db:"actorName" json:"actorName"`
	Before    json.RawMessage `db:"before" json:"before"`
	Exists    bool            `db:"exists" json:"restored"`
}

// GET /api/owner/deleted?entity=&limit=
func (h *DeletedRecordsHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	ent := r.URL.Query().Get("entity")
	if ent != "" {
		if _, ok := RestorableEntities[ent]; !ok {
			WriteError(w, http.StatusBadRequest, "نوع غير معروف")
			return
		}
	}
	rows := []deletedRow{}
	if err := h.db.Select(&rows, `SELECT a.id, a.at, a.entity, a."entityId", e.name AS "actorName", a.before, false AS "exists"
		FROM "AuditTrail" a LEFT JOIN "Employee" e ON e.id = a."actorId"
		WHERE a.action = 'DELETE' AND ($1 = '' OR a.entity = $1) ORDER BY a.id DESC LIMIT $2`, ent, limit); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	// (11-04 — ٧ن) «راجع؟» باستعلام واحد لكل نوع، مو واحد لكل سطر
	idsBy := map[string][]string{}
	for _, x := range rows {
		if _, ok := RestorableEntities[x.Entity]; ok && x.EntityID != nil {
			idsBy[x.Entity] = append(idsBy[x.Entity], *x.EntityID)
		}
	}
	back := map[string]bool{}
	for ent, ids := range idsBy {
		found := []string{}
		errlog.Err(h.db.Select(&found, `SELECT id::text FROM "`+ent+`" WHERE id::text = ANY($1)`, pq.Array(ids)))
		for _, id := range found {
			back[ent+"|"+id] = true
		}
	}
	out := rows[:0]
	for _, x := range rows {
		lbl, ok := RestorableEntities[x.Entity]
		if !ok {
			continue
		}
		x.Label = lbl
		x.Exists = x.EntityID != nil && back[x.Entity+"|"+*x.EntityID]
		out = append(out, x)
	}
	WriteJSON(w, http.StatusOK, map[string]any{"rows": out, "entities": RestorableEntities})
}

// POST /api/owner/deleted/{id}/restore — يرجّع الصف كما كان لحظة الحذف.
func (h *DeletedRecordsHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "رقم غير صالح")
		return
	}
	var a struct {
		Entity   string          `db:"entity"`
		EntityID *string         `db:"entityId"`
		Before   json.RawMessage `db:"before"`
	}
	if err := h.db.Get(&a, `SELECT entity, "entityId", before FROM "AuditTrail" WHERE id = $1 AND action = 'DELETE'`, id); err != nil {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	if _, ok := RestorableEntities[a.Entity]; !ok || a.EntityID == nil {
		WriteError(w, http.StatusBadRequest, "هذا النوع ما يرجع")
		return
	}
	// اسم الجدول من القائمة الثابتة فوگ بس — ماكو إدخال من المستخدم بالـ SQL
	tx, err := h.db.Beginx()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	defer tx.Rollback()
	errlog.Exec(tx.Exec(`SELECT set_config('app.actor', $1, true)`, middleware.EmployeeIDFromContext(r)))
	var exists bool
	errlog.Err(tx.Get(&exists, `SELECT EXISTS (SELECT 1 FROM "`+a.Entity+`" WHERE id::text = $1)`, *a.EntityID))
	if exists {
		WriteError(w, http.StatusConflict, "هذا السجل راجع أصلاً")
		return
	}
	if _, err := tx.Exec(`INSERT INTO "`+a.Entity+`" SELECT * FROM jsonb_populate_record(NULL::"`+a.Entity+`", $1::jsonb)`, string(a.Before)); err != nil {
		WriteError(w, http.StatusBadRequest, "ما رجع — يمكن شي مربوط بيه انحذف هم (مثل المشروع الأب): "+err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
