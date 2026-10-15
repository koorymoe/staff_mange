package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ علاقات الموردين — قرار (ع) 10-15 ═══
// «الموردين لازم ينربطون بالعلاقات العامة وبوحدة التقنيين… نظام تذكيري:
// صارلنا شكد ما متواصلين وياهم، وآخر الحركات وياهم شوكت وشنو».
//   - موردين المشتريات (Supplier) وموردين التقنيين (TechSupplier) بشاشة وحدة.
//   - كل تواصل ينسجّل (اتصال · زيارة · رسالة · اجتماع · ملاحظة) بمنو ومتى.
//   - المشتريات منهم تنحسب حركات تلقائياً.
//   - مورد عبر الحد (افتراضي ٣٠ يوم) بلا أي حركة → تذكير يومي للعلاقات العامة والمدير.

type SupplierRelationsHandler struct {
	db     *sqlx.DB
	notify func(string)
}

func NewSupplierRelationsHandler(db *sqlx.DB, notify func(string)) *SupplierRelationsHandler {
	return &SupplierRelationsHandler{db: db, notify: notify}
}

var contactKinds = map[string]string{"CALL": "📞 اتصال", "VISIT": "🚶 زيارة", "MESSAGE": "💬 رسالة", "MEETING": "🤝 اجتماع", "NOTE": "📝 ملاحظة"}

func (h *SupplierRelationsHandler) contactDays() int {
	var d int
	if err := h.db.Get(&d, `SELECT "contactDays" FROM "PurchasePolicy" WHERE id = 1`); err != nil || d <= 0 {
		return 30
	}
	return d
}

type relationRow struct {
	Kind          string     `db:"kind" json:"kind"`
	ID            string     `db:"id" json:"id"`
	Name          string     `db:"name" json:"name"`
	Phone         *string    `db:"phone" json:"phone"`
	Specialty     *string    `db:"specialty" json:"specialty"`
	LastContactAt *time.Time `db:"lastContactAt" json:"lastContactAt"`
	LastContact   *string    `db:"lastContact" json:"lastContact"`
	LastContactBy *string    `db:"lastContactBy" json:"lastContactBy"`
	LastBuyAt     *time.Time `db:"lastBuyAt" json:"lastBuyAt"`
	LastBuy       *string    `db:"lastBuy" json:"lastBuy"`
	Contacts90    int        `db:"contacts90" json:"contacts90"`
	Buys90        int        `db:"buys90" json:"buys90"`
	DaysSince     *int       `db:"-" json:"daysSince"`
	Overdue       bool       `db:"-" json:"overdue"`
}

const relationsSQL = `
	WITH sup AS (
		SELECT 'PROC' AS kind, s.id, s."companyName" AS name, s.phone,
			(SELECT string_agg(sp.name, '، ') FROM "SupplierSpecialtyLink" l JOIN "SupplierSpecialty" sp ON sp.id = l."specialtyId" WHERE l."supplierId" = s.id) AS specialty
		FROM "Supplier" s
		UNION ALL
		SELECT 'TECH', t.id, t."companyName", t.phone, t.specialty FROM "TechSupplier" t
	)
	SELECT sup.*,
		c.at AS "lastContactAt", c.txt AS "lastContact", c.by AS "lastContactBy",
		b.at AS "lastBuyAt", b.txt AS "lastBuy",
		(SELECT count(*) FROM "SupplierContact" x WHERE x."supplierKind" = sup.kind AND x."supplierId" = sup.id AND x.at > now() - interval '90 days')::int AS contacts90,
		(SELECT count(DISTINCT "refId") FROM "PurchaseLine" p WHERE p.source = sup.kind AND p."supplierId" = sup.id AND p.at > now() - interval '90 days')::int AS buys90
	FROM sup
	LEFT JOIN LATERAL (SELECT x.at, x.kind || COALESCE(': ' || x.note, '') AS txt, e.name AS by FROM "SupplierContact" x LEFT JOIN "Employee" e ON e.id = x."byId"
		WHERE x."supplierKind" = sup.kind AND x."supplierId" = sup.id ORDER BY x.at DESC LIMIT 1) c ON true
	LEFT JOIN LATERAL (SELECT p.at, p.product || ' ×' || p.qty AS txt FROM "PurchaseLine" p
		WHERE p.source = sup.kind AND p."supplierId" = sup.id ORDER BY p.at DESC LIMIT 1) b ON true`

func (h *SupplierRelationsHandler) rows() ([]relationRow, error) {
	rows := []relationRow{}
	if err := h.db.Select(&rows, relationsSQL); err != nil {
		return nil, err
	}
	limit := h.contactDays()
	now := time.Now()
	for i := range rows {
		r := &rows[i]
		last := r.LastContactAt
		if r.LastBuyAt != nil && (last == nil || r.LastBuyAt.After(*last)) {
			last = r.LastBuyAt
		}
		if last != nil {
			d := int(now.Sub(*last).Hours() / 24)
			r.DaysSince = &d
			r.Overdue = d > limit
		} else {
			r.Overdue = true
		}
		if r.LastContact != nil {
			for k, v := range contactKinds {
				if strings.HasPrefix(*r.LastContact, k) {
					t := v + strings.TrimPrefix(*r.LastContact, k)
					r.LastContact = &t
				}
			}
		}
	}
	return rows, nil
}

// GET /api/supplier-relations
func (h *SupplierRelationsHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.rows()
	if err != nil {
		log.Printf("[supplier-relations] %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"rows": rows, "contactDays": h.contactDays(), "kinds": contactKinds})
}

func validKind(k string) bool { return k == "PROC" || k == "TECH" }

// POST /api/supplier-relations/{kind}/{id}/contact {type, note}
func (h *SupplierRelationsHandler) AddContact(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Type string `json:"type"`
		Note string `json:"note"`
	}
	if err := DecodeJSON(r, &b); err != nil || !validKind(r.PathValue("kind")) {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	if _, ok := contactKinds[b.Type]; !ok {
		WriteError(w, http.StatusBadRequest, "نوع التواصل غير معروف")
		return
	}
	var note *string
	if n := strings.TrimSpace(b.Note); n != "" {
		note = &n
	}
	if _, err := h.db.Exec(`INSERT INTO "SupplierContact" (id, "supplierKind", "supplierId", kind, note, "byId") VALUES ($1,$2,$3,$4,$5,$6)`,
		uuid.NewString(), r.PathValue("kind"), r.PathValue("id"), b.Type, note, middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GET /api/supplier-relations/{kind}/{id}/timeline — كل الحركات (تواصل + مشتريات)
func (h *SupplierRelationsHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	kind, id := r.PathValue("kind"), r.PathValue("id")
	if !validKind(kind) {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	type ev struct {
		At   time.Time `db:"at" json:"at"`
		Type string    `db:"type" json:"type"`
		Text *string   `db:"text" json:"text"`
		By   *string   `db:"by" json:"by"`
	}
	rows := []ev{}
	if err := h.db.Select(&rows, `SELECT * FROM (
			SELECT x.at, x.kind AS type, x.note AS text, e.name AS by FROM "SupplierContact" x LEFT JOIN "Employee" e ON e.id = x."byId"
			WHERE x."supplierKind" = $1 AND x."supplierId" = $2
			UNION ALL
			SELECT p.at, 'PURCHASE', p.code || ' · ' || p.product || ' ×' || p.qty || ' بـ' || round(p.price)::text || ' د.ع', e.name
			FROM "PurchaseLine" p LEFT JOIN "Employee" e ON e.id = p."buyerId" WHERE p.source = $1 AND p."supplierId" = $2
		) t ORDER BY at DESC LIMIT 200`, kind, id); err != nil {
		log.Printf("[supplier-relations] timeline: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/supplier-relations/policy {contactDays}
func (h *SupplierRelationsHandler) SetPolicy(w http.ResponseWriter, r *http.Request) {
	var b struct {
		ContactDays int `json:"contactDays"`
	}
	if err := DecodeJSON(r, &b); err != nil || b.ContactDays < 1 || b.ContactDays > 365 {
		WriteError(w, http.StatusBadRequest, "عدد أيام غير صالح")
		return
	}
	_, _ = h.db.Exec(`UPDATE "PurchasePolicy" SET "contactDays" = $1 WHERE id = 1`, b.ContactDays)
	WriteJSON(w, http.StatusOK, b)
}

// RunReminders تذكير يومي (الساعة ١٠ صباحاً بغداد): الموردين الي عبروا الحد بلا أي حركة.
func (h *SupplierRelationsHandler) RunReminders() {
	lastDay := ""
	for {
		now := time.Now().In(baghdadLoc())
		if day := now.Format("2006-01-02"); now.Hour() >= 10 && day != lastDay {
			lastDay = day
			if rows, err := h.rows(); err == nil && h.notify != nil {
				names := []string{}
				for _, r := range rows {
					if r.Overdue {
						tag := "مشتريات"
						if r.Kind == "TECH" {
							tag = "تقنيين"
						}
						since := "ماكو ولا تواصل"
						if r.DaysSince != nil {
							since = fmt.Sprintf("%d يوم", *r.DaysSince)
						}
						names = append(names, fmt.Sprintf("%s (%s، %s)", r.Name, tag, since))
					}
				}
				if len(names) > 0 {
					more := ""
					if len(names) > 8 {
						more = fmt.Sprintf(" و%d غيرهم", len(names)-8)
						names = names[:8]
					}
					h.notify(fmt.Sprintf("🤝 تذكير علاقات الموردين: %d مورد صارلنا أكثر من %d يوم ما تواصلنا وياهم — %s%s",
						len(names), h.contactDays(), strings.Join(names, " · "), more))
				}
			}
		}
		time.Sleep(30 * time.Minute)
	}
}
