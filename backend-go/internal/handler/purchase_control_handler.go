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

// ═══ رقابة المشتريات — قرار (ع) 10-15 ═══
// «اريد رقابة قوية قوية على المشتريات»:
//   - المشتري يسجّل الشراء (المورد والسعر) والمحاسب يثبّته.
//   - بعد كل شراء ينطلب من المشتري تقييم المورد ببنود — أولها السعر.
//   - سجل أسعار لكل مادة: منو أرخص ومنو أغلى.
//   - ماتركس ينبّه: اشترى من الأغلى (والمادة موجودة أرخص عند غيره)، أو المورد رفع سعره.

type PurchaseControlHandler struct {
	db     *sqlx.DB
	notify func(message string)
}

func NewPurchaseControlHandler(db *sqlx.DB, notify func(string)) *PurchaseControlHandler {
	return &PurchaseControlHandler{db: db, notify: notify}
}

const (
	priceWindow   = "180 days"
	pricierRatio  = 1.10 // أغلى من أرخص مورد بأكثر من ١٠٪
	priceJumpRate = 1.15 // نفس المورد رفع سعره أكثر من ١٥٪
)

type pricedItem struct {
	ID         string   `db:"id"`
	Name       string   `db:"productName"`
	Key        string   `db:"key"`
	Qty        int      `db:"quantity"`
	UnitPrice  *float64 `db:"unitPrice"`
	SupplierID *string  `db:"supplierId"`
	BuyerID    *string  `db:"fulfilledById"`
	Code       string   `db:"code"`
}

// CheckPrices بعد التجهيز (الشراء): يقارن سعر كل مادة بأسعار الموردين الثانيين.
func (h *PurchaseControlHandler) CheckPrices(procurementID string) {
	var items []pricedItem
	if err := h.db.Select(&items, `SELECT i.id, i."productName", product_key(i."productName", i.model) AS key, i.quantity, i."unitPrice",
		r."supplierId", r."fulfilledById", r.code FROM "ProcurementItem" i JOIN "ProcurementRequest" r ON r.id = i."requestId"
		WHERE r.id = $1 AND i."unitPrice" IS NOT NULL AND i."unitPrice" > 0`, procurementID); err != nil {
		log.Printf("[purchase-control] %v", err)
		return
	}
	for _, it := range items {
		if it.SupplierID == nil {
			continue
		}
		// أرخص مورد ثاني لنفس المادة (آخر سعر لكل مورد)
		var best struct {
			SupplierID string  `db:"supplierId"`
			Name       string  `db:"name"`
			Price      float64 `db:"price"`
		}
		// آخر سعر لكل مورد ثاني، ونختار الأرخص بيناتهم
		err := h.db.Get(&best, `SELECT * FROM (SELECT DISTINCT ON (r."supplierId") r."supplierId", s."companyName" AS name, i."unitPrice" AS price
				FROM "ProcurementItem" i JOIN "ProcurementRequest" r ON r.id = i."requestId" JOIN "Supplier" s ON s.id = r."supplierId"
				WHERE product_key(i."productName", i.model) = $1 AND r."supplierId" <> $2 AND i."unitPrice" > 0
				  AND r."fulfilledAt" > now() - interval '`+priceWindow+`' AND i.id <> $3
				ORDER BY r."supplierId", r."fulfilledAt" DESC) x ORDER BY price LIMIT 1`, it.Key, *it.SupplierID, it.ID)
		if err == nil {
			if *it.UnitPrice > best.Price*pricierRatio {
				loss := (*it.UnitPrice - best.Price) * float64(it.Qty)
				h.alert(procurementID, it, "PRICIER", &best.SupplierID, best.Price, loss,
					fmt.Sprintf("🚨 مشتريات %s: «%s» انشترت بـ%s د.ع والقطعة موجودة عند %s بـ%s — خسارة %s د.ع",
						it.Code, it.Name, money(*it.UnitPrice), best.Name, money(best.Price), money(loss)))
			}
		}
		// نفس المورد رفع سعره عن آخر مرة
		var last float64
		if err := h.db.Get(&last, `SELECT i."unitPrice" FROM "ProcurementItem" i JOIN "ProcurementRequest" r ON r.id = i."requestId"
			WHERE product_key(i."productName", i.model) = $1 AND r."supplierId" = $2 AND i."unitPrice" > 0 AND i.id <> $3
			  AND r."fulfilledAt" IS NOT NULL ORDER BY r."fulfilledAt" DESC LIMIT 1`, it.Key, *it.SupplierID, it.ID); err == nil && last > 0 && *it.UnitPrice > last*priceJumpRate {
			loss := (*it.UnitPrice - last) * float64(it.Qty)
			h.alert(procurementID, it, "JUMP", it.SupplierID, last, loss,
				fmt.Sprintf("📈 مشتريات %s: المورد رفع سعر «%s» من %s إلى %s د.ع (+%.0f٪)",
					it.Code, it.Name, money(last), money(*it.UnitPrice), (*it.UnitPrice/last-1)*100))
		}
	}
}

func (h *PurchaseControlHandler) alert(procID string, it pricedItem, kind string, refSup *string, refPrice, loss float64, msg string) {
	r, err := h.db.Exec(`INSERT INTO "ProcurementPriceAlert" (id, "procurementId", "itemId", kind, "productName", "supplierId", "unitPrice", quantity,
		"refSupplierId", "refPrice", "lossIqd", "buyerId") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT DO NOTHING`,
		uuid.NewString(), procID, it.ID, kind, it.Name, it.SupplierID, *it.UnitPrice, it.Qty, refSup, refPrice, loss, it.BuyerID)
	if err != nil {
		log.Printf("[purchase-control] alert: %v", err)
		return
	}
	if n, _ := r.RowsAffected(); n > 0 && h.notify != nil {
		h.notify(msg)
	}
}

func money(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	var out []string
	for len(s) > 3 {
		out = append([]string{s[len(s)-3:]}, out...)
		s = s[:len(s)-3]
	}
	return strings.Join(append([]string{s}, out...), ",")
}

// POST /api/procurement/{id}/review — تقييم المورد ببنود (١-٥): السعر · الجودة · السرعة · الالتزام · التعامل
func (h *PurchaseControlHandler) Review(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Price, Quality, Speed, Punctuality, Dealing int
		Note                                        *string
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	for _, v := range []int{b.Price, b.Quality, b.Speed, b.Punctuality, b.Dealing} {
		if v < 1 || v > 5 {
			WriteError(w, http.StatusBadRequest, "كل بند من ١ إلى ٥")
			return
		}
	}
	id := r.PathValue("id")
	var sup *string
	if err := h.db.Get(&sup, `SELECT "supplierId" FROM "ProcurementRequest" WHERE id = $1`, id); err != nil || sup == nil {
		WriteError(w, http.StatusBadRequest, "الطلب ما بيه مورد")
		return
	}
	if _, err := h.db.Exec(`INSERT INTO "SupplierReview" (id, "supplierId", "procurementId", price, quality, speed, punctuality, dealing, note, "ratedById")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT ("procurementId", "ratedById") WHERE "procurementId" IS NOT NULL
		DO UPDATE SET price = $4, quality = $5, speed = $6, punctuality = $7, dealing = $8, note = $9`,
		uuid.NewString(), *sup, id, b.Price, b.Quality, b.Speed, b.Punctuality, b.Dealing, b.Note, middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ التقييم")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// PUT /api/procurement/{id}/verify — المحاسب يثبّت الشراء (المورد والسعر)
func (h *PurchaseControlHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var b struct{ Note *string }
	_ = DecodeJSON(r, &b)
	id := r.PathValue("id")
	var fulfilled bool
	if err := h.db.Get(&fulfilled, `SELECT "fulfilledAt" IS NOT NULL FROM "ProcurementRequest" WHERE id = $1`, id); err != nil || !fulfilled {
		WriteError(w, http.StatusBadRequest, "الطلب لسه ما انشترى")
		return
	}
	if _, err := h.db.Exec(`INSERT INTO "ProcurementVerify" ("procurementId", "verifiedById", note) VALUES ($1,$2,$3)
		ON CONFLICT ("procurementId") DO UPDATE SET "verifiedById" = $2, note = $3, "verifiedAt" = now()`, id, middleware.EmployeeIDFromContext(r), b.Note); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر التثبيت")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GET /api/procurement/control — حالة كل شراء (تقييمي، التثبيت، التنبيهات) لشاشة المشتريات
func (h *PurchaseControlHandler) Control(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	type row struct {
		ID         string     `db:"id" json:"id"`
		VerifiedAt *time.Time `db:"verifiedAt" json:"verifiedAt"`
		VerifiedBy *string    `db:"verifiedBy" json:"verifiedBy"`
		MyReview   bool       `db:"myReview" json:"myReview"`
		Alerts     int        `db:"alerts" json:"alerts"`
	}
	rows := []row{}
	_ = h.db.Select(&rows, `SELECT r.id, v."verifiedAt", e.name AS "verifiedBy",
		EXISTS (SELECT 1 FROM "SupplierReview" x WHERE x."procurementId" = r.id AND x."ratedById" = $1) AS "myReview",
		(SELECT count(*) FROM "ProcurementPriceAlert" a WHERE a."procurementId" = r.id)::int AS alerts
		FROM "ProcurementRequest" r LEFT JOIN "ProcurementVerify" v ON v."procurementId" = r.id LEFT JOIN "Employee" e ON e.id = v."verifiedById"
		WHERE r."fulfilledAt" > now() - interval '120 days'`, me)
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/procurement/prices?q= — سجل الأسعار: لكل مادة، كل مورد بآخر سعر وأرخص سعر وعدد المرات
func (h *PurchaseControlHandler) Prices(w http.ResponseWriter, r *http.Request) {
	q := "%" + strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q"))) + "%"
	type row struct {
		Key       string    `db:"key" json:"key"`
		Product   string    `db:"product" json:"product"`
		Supplier  string    `db:"supplier" json:"supplier"`
		LastPrice float64   `db:"lastPrice" json:"lastPrice"`
		MinPrice  float64   `db:"minPrice" json:"minPrice"`
		Times     int       `db:"times" json:"times"`
		LastAt    time.Time `db:"lastAt" json:"lastAt"`
	}
	rows := []row{}
	if err := h.db.Select(&rows, `SELECT key, max(product) AS product, supplier,
		(array_agg(price ORDER BY at DESC))[1] AS "lastPrice", min(price) AS "minPrice", count(*)::int AS times, max(at) AS "lastAt"
		FROM (SELECT product_key(i."productName", i.model) AS key, i."productName" AS product, s."companyName" AS supplier,
			i."unitPrice" AS price, r."fulfilledAt" AS at
			FROM "ProcurementItem" i JOIN "ProcurementRequest" r ON r.id = i."requestId" JOIN "Supplier" s ON s.id = r."supplierId"
			WHERE i."unitPrice" > 0 AND r."fulfilledAt" > now() - interval '365 days') x
		WHERE key LIKE $1 GROUP BY key, supplier ORDER BY key, "lastPrice" LIMIT 1000`, q); err != nil {
		log.Printf("[purchase-control] prices: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأسعار")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/procurement/alerts — تنبيهات ماتركس على المشتريات (للمالك والمدير والمراقب والمحاسب)
func (h *PurchaseControlHandler) Alerts(w http.ResponseWriter, r *http.Request) {
	type row struct {
		ID          string     `db:"id" json:"id"`
		Kind        string     `db:"kind" json:"kind"`
		Code        string     `db:"code" json:"code"`
		Product     string     `db:"productName" json:"product"`
		Supplier    *string    `db:"supplier" json:"supplier"`
		UnitPrice   float64    `db:"unitPrice" json:"unitPrice"`
		Quantity    int        `db:"quantity" json:"quantity"`
		RefSupplier *string    `db:"refSupplier" json:"refSupplier"`
		RefPrice    float64    `db:"refPrice" json:"refPrice"`
		Loss        float64    `db:"lossIqd" json:"loss"`
		Buyer       *string    `db:"buyer" json:"buyer"`
		CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
		ReviewedAt  *time.Time `db:"reviewedAt" json:"reviewedAt"`
		ReviewNote  *string    `db:"reviewNote" json:"reviewNote"`
	}
	rows := []row{}
	_ = h.db.Select(&rows, `SELECT a.id, a.kind, r.code, a."productName", s."companyName" AS supplier, a."unitPrice", a.quantity,
		rs."companyName" AS "refSupplier", a."refPrice", a."lossIqd", e.name AS buyer, a."createdAt", a."reviewedAt", a."reviewNote"
		FROM "ProcurementPriceAlert" a JOIN "ProcurementRequest" r ON r.id = a."procurementId"
		LEFT JOIN "Supplier" s ON s.id = a."supplierId" LEFT JOIN "Supplier" rs ON rs.id = a."refSupplierId"
		LEFT JOIN "Employee" e ON e.id = a."buyerId" ORDER BY a."reviewedAt" IS NOT NULL, a."createdAt" DESC LIMIT 300`)
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/procurement-alerts/{id} — المالك/المدير يحسم التنبيه بملاحظة (مثلاً «الأرخص ما چان متوفر»)
func (h *PurchaseControlHandler) ReviewAlert(w http.ResponseWriter, r *http.Request) {
	var b struct{ Note string }
	_ = DecodeJSON(r, &b)
	if strings.TrimSpace(b.Note) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب الملاحظة")
		return
	}
	_, _ = h.db.Exec(`UPDATE "ProcurementPriceAlert" SET "reviewedAt" = now(), "reviewNote" = $2 WHERE id = $1`, r.PathValue("id"), strings.TrimSpace(b.Note))
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GET /api/suppliers/scores — معدل كل بند لكل مورد (من التقييمات ببنود)
func (h *PurchaseControlHandler) SupplierScores(w http.ResponseWriter, r *http.Request) {
	type row struct {
		SupplierID  string  `db:"supplierId" json:"supplierId"`
		N           int     `db:"n" json:"n"`
		Price       float64 `db:"price" json:"price"`
		Quality     float64 `db:"quality" json:"quality"`
		Speed       float64 `db:"speed" json:"speed"`
		Punctuality float64 `db:"punctuality" json:"punctuality"`
		Dealing     float64 `db:"dealing" json:"dealing"`
		Alerts      int     `db:"alerts" json:"alerts"`
	}
	rows := []row{}
	_ = h.db.Select(&rows, `SELECT s.id AS "supplierId", count(v.id)::int AS n, COALESCE(avg(v.price),0) AS price, COALESCE(avg(v.quality),0) AS quality,
		COALESCE(avg(v.speed),0) AS speed, COALESCE(avg(v.punctuality),0) AS punctuality, COALESCE(avg(v.dealing),0) AS dealing,
		(SELECT count(*) FROM "ProcurementPriceAlert" a WHERE a."supplierId" = s.id)::int AS alerts
		FROM "Supplier" s LEFT JOIN "SupplierReview" v ON v."supplierId" = s.id GROUP BY s.id`)
	WriteJSON(w, http.StatusOK, rows)
}
