package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
)

// ═══ الأسعار الستاندرد — قرار (ع) 10-15 ═══
// «لازم نخلي فد أسعار ستاندرد… بين فترة وفترة تتغيّر حسب تقييم الأسعار».
//   - كل مادة إلها سعر مرجعي: تلقائي (أقل آخر سعر بين الموردين، آخر ٩٠ يوم)
//     أو يدوي يثبّته المدير ويغلب.
//   - ماتركس يقترح تعديل لما السوق يتغيّر (أكثر من ٥٪) والمدير يوافق بضغطة.
//   - أي شراء (موظف المشتريات أو الفني من مورده) فوق المرجعي + السقف (افتراضي ١٠٪)
//     → تنبيه للمدير والمراقب.

func (h *PurchaseControlHandler) ceiling() float64 {
	var c float64
	if err := h.db.Get(&c, `SELECT "ceilingPct" FROM "PurchasePolicy" WHERE id = 1`); err != nil {
		return 10
	}
	return c
}

type marketRow struct {
	Key      string    `db:"key"`
	Product  string    `db:"product"`
	Price    float64   `db:"price"`
	Supplier string    `db:"supplier"`
	N        int       `db:"n"`
	Last     time.Time `db:"last"`
}

// RefreshStandards يحسب سعر السوق لكل مادة: أقل «آخر سعر» بين الموردين خلال ٩٠ يوم
// (مستبعد الموردين الي تقييم سعرهم ضعيف < ٢.٥). مادة بلا مرجعي → ينحط تلقائياً،
// ومرجعي موجود ويفرق أكثر من ٥٪ → اقتراح ينتظر موافقة المدير.
func (h *PurchaseControlHandler) RefreshStandards() {
	rows := []marketRow{}
	if err := h.db.Select(&rows, `WITH last AS (
			SELECT DISTINCT ON (key, supplier) key, product, price, supplier, "supplierId", at FROM "PurchaseLine"
			WHERE at > now() - interval '90 days' ORDER BY key, supplier, at DESC
		), ok AS (
			SELECT l.* FROM last l WHERE COALESCE((SELECT avg(v.price) FROM "SupplierReview" v WHERE v."supplierId" = l."supplierId"), 5) >= 2.5
		)
		SELECT DISTINCT ON (key) key, product, price, supplier,
			(SELECT count(*) FROM "PurchaseLine" x WHERE x.key = ok.key AND x.at > now() - interval '90 days')::int AS n, at AS last
		FROM ok ORDER BY key, price`); err != nil {
		log.Printf("[standard-price] %v", err)
		return
	}
	for _, m := range rows {
		var cur struct {
			Price  float64 `db:"price"`
			Source string  `db:"source"`
		}
		if err := h.db.Get(&cur, `SELECT price, source FROM "StandardPrice" WHERE key = $1`, m.Key); err != nil {
			_, _ = h.db.Exec(`INSERT INTO "StandardPrice" (key, product, price, source) VALUES ($1,$2,$3,'AUTO') ON CONFLICT DO NOTHING`, m.Key, m.Product, m.Price)
			continue
		}
		if cur.Price > 0 && (m.Price > cur.Price*1.05 || m.Price < cur.Price*0.95) {
			basis := fmt.Sprintf("أرخص سعر بالسوق هسه %s د.ع عند %s (%d شراء بآخر ٩٠ يوم)", money(m.Price), m.Supplier, m.N)
			_, _ = h.db.Exec(`INSERT INTO "StandardPriceSuggestion" (key, product, current, suggested, basis) VALUES ($1,$2,$3,$4,$5)
				ON CONFLICT (key) DO UPDATE SET current = $3, suggested = $4, basis = $5, "createdAt" = now()`, m.Key, m.Product, cur.Price, m.Price, basis)
		} else {
			_, _ = h.db.Exec(`DELETE FROM "StandardPriceSuggestion" WHERE key = $1`, m.Key)
		}
	}
}

// RunStandardsLoop كل ٦ ساعات يحدّث الاقتراحات
func (h *PurchaseControlHandler) RunStandardsLoop() {
	for {
		h.RefreshStandards()
		time.Sleep(6 * time.Hour)
	}
}

// checkStandard يقارن سعر شراء بالمرجعي. يرجّع true إذا انعمل تنبيه.
func (h *PurchaseControlHandler) checkStandard(procID, expenseID, techSup *string, it pricedItem, who string) {
	var std float64
	if err := h.db.Get(&std, `SELECT price FROM "StandardPrice" WHERE key = $1`, it.Key); err != nil || std <= 0 {
		return
	}
	ceil := h.ceiling()
	if *it.UnitPrice <= std*(1+ceil/100) {
		return
	}
	loss := (*it.UnitPrice - std) * float64(it.Qty)
	pct := (*it.UnitPrice/std - 1) * 100
	msg := fmt.Sprintf("📏 %s: «%s» انشترت بـ%s د.ع والسعر الستاندرد %s (+%.0f٪) — زيادة %s د.ع",
		who, it.Name, money(*it.UnitPrice), money(std), pct, money(loss))
	r, err := h.db.Exec(`INSERT INTO "ProcurementPriceAlert" (id, "procurementId", "expenseId", "itemId", kind, "productName", "supplierId", "techSupplierId",
		"unitPrice", quantity, "refPrice", "lossIqd", "buyerId") VALUES ($1,$2,$3,$4,'OVER_STD',$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT DO NOTHING`,
		uuid.NewString(), procID, expenseID, it.ID, it.Name, it.SupplierID, techSup, *it.UnitPrice, it.Qty, std, loss, it.BuyerID)
	if err != nil {
		log.Printf("[standard-price] alert: %v", err)
		return
	}
	if n, _ := r.RowsAffected(); n > 0 && h.notify != nil {
		h.notify(msg)
	}
}

// AfterExpensePurchase شراء الفني من مورده: ينحفظ وينقارن بالمرجعي.
func (h *PurchaseControlHandler) AfterExpensePurchase(expenseID, employeeID string, p model.ExpensePurchaseInput) {
	if _, err := h.db.Exec(`INSERT INTO "ExpensePurchase" ("expenseId", "techSupplierId", "productName", quantity, "unitPrice") VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT DO NOTHING`, expenseID, p.TechSupplierID, p.ProductName, p.Quantity, p.UnitPrice); err != nil {
		log.Printf("[standard-price] expense purchase: %v", err)
		return
	}
	var key, name string
	_ = h.db.Get(&key, `SELECT product_key($1, NULL)`, p.ProductName)
	_ = h.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, employeeID)
	price := p.UnitPrice
	who := "مصروف الفني " + name
	if p.TechSupplierID == nil || *p.TechSupplierID == "" {
		who += " (مورد غير مسجّل بقائمتنا)"
		if h.notify != nil {
			h.notify(fmt.Sprintf("⚠️ %s اشترى «%s» من مورد مو مسجّل بقائمة موردين الفنيين", name, p.ProductName))
		}
	}
	h.checkStandard(nil, &expenseID, p.TechSupplierID, pricedItem{ID: "exp:" + expenseID, Name: p.ProductName, Key: key, Qty: p.Quantity, UnitPrice: &price, BuyerID: &employeeID}, who)
}

// ─── الشاشة ───

// GET /api/procurement-standards — المرجعي + الاقتراحات + مواد بلا مرجعي + السقف
func (h *PurchaseControlHandler) Standards(w http.ResponseWriter, r *http.Request) {
	type std struct {
		Key       string    `db:"key" json:"key"`
		Product   string    `db:"product" json:"product"`
		Price     float64   `db:"price" json:"price"`
		Source    string    `db:"source" json:"source"`
		SetBy     *string   `db:"setBy" json:"setBy"`
		UpdatedAt time.Time `db:"updatedAt" json:"updatedAt"`
		Market    *float64  `db:"market" json:"market"`
	}
	type sug struct {
		Key       string  `db:"key" json:"key"`
		Product   string  `db:"product" json:"product"`
		Current   float64 `db:"current" json:"current"`
		Suggested float64 `db:"suggested" json:"suggested"`
		Basis     string  `db:"basis" json:"basis"`
	}
	stds, sugs := []std{}, []sug{}
	_ = h.db.Select(&stds, `SELECT s.key, s.product, s.price, s.source, e.name AS "setBy", s."updatedAt",
		(SELECT min(price) FROM "PurchaseLine" l WHERE l.key = s.key AND l.at > now() - interval '90 days') AS market
		FROM "StandardPrice" s LEFT JOIN "Employee" e ON e.id = s."setById" ORDER BY s.product`)
	_ = h.db.Select(&sugs, `SELECT key, product, current, suggested, basis FROM "StandardPriceSuggestion" ORDER BY "createdAt" DESC`)
	WriteJSON(w, http.StatusOK, map[string]any{"standards": stds, "suggestions": sugs, "ceilingPct": h.ceiling()})
}

// PUT /api/procurement-standards — المدير يثبّت سعر مرجعي يدوي {product, price}
func (h *PurchaseControlHandler) SetStandard(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Product string  `json:"product"`
		Price   float64 `json:"price"`
	}
	if err := DecodeJSON(r, &b); err != nil || strings.TrimSpace(b.Product) == "" || b.Price <= 0 {
		WriteError(w, http.StatusBadRequest, "اكتب المادة والسعر")
		return
	}
	var key string
	_ = h.db.Get(&key, `SELECT product_key($1, NULL)`, b.Product)
	if _, err := h.db.Exec(`INSERT INTO "StandardPrice" (key, product, price, source, "setById") VALUES ($1,$2,$3,'MANUAL',$4)
		ON CONFLICT (key) DO UPDATE SET price = $3, source = 'MANUAL', "setById" = $4, "updatedAt" = now()`,
		key, strings.TrimSpace(b.Product), b.Price, middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	_, _ = h.db.Exec(`DELETE FROM "StandardPriceSuggestion" WHERE key = $1`, key)
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /api/procurement-standards/accept {key, accept:bool} — موافقة/رفض اقتراح ماتركس
func (h *PurchaseControlHandler) AcceptSuggestion(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Key    string `json:"key"`
		Accept bool   `json:"accept"`
	}
	if err := DecodeJSON(r, &b); err != nil || b.Key == "" {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	if b.Accept {
		_, _ = h.db.Exec(`UPDATE "StandardPrice" s SET price = x.suggested, source = 'MANUAL', "setById" = $2, "updatedAt" = now()
			FROM "StandardPriceSuggestion" x WHERE x.key = s.key AND s.key = $1`, b.Key, middleware.EmployeeIDFromContext(r))
	}
	_, _ = h.db.Exec(`DELETE FROM "StandardPriceSuggestion" WHERE key = $1`, b.Key)
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// PUT /api/procurement-policy/ceiling {ceilingPct}
func (h *PurchaseControlHandler) SetCeiling(w http.ResponseWriter, r *http.Request) {
	var b struct {
		CeilingPct float64 `json:"ceilingPct"`
	}
	if err := DecodeJSON(r, &b); err != nil || b.CeilingPct < 0 || b.CeilingPct > 100 {
		WriteError(w, http.StatusBadRequest, "نسبة غير صالحة")
		return
	}
	_, _ = h.db.Exec(`UPDATE "PurchasePolicy" SET "ceilingPct" = $1, "updatedAt" = now() WHERE id = 1`, b.CeilingPct)
	WriteJSON(w, http.StatusOK, b)
}
