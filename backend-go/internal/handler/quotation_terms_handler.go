package handler

import (
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// QuotationTermsHandler — شروط وأحكام عرض السعر (قرار (ع) 10-08).
type QuotationTermsHandler struct{ db *sqlx.DB }

func NewQuotationTermsHandler(db *sqlx.DB) *QuotationTermsHandler { return &QuotationTermsHandler{db: db} }

type quotationTerm struct {
	ID        string `db:"id" json:"id"`
	Text      string `db:"text" json:"text"`
	SortOrder int    `db:"sortOrder" json:"sortOrder"`
}

func (h *QuotationTermsHandler) list(w http.ResponseWriter) {
	rows := []quotationTerm{}
	if err := h.db.Select(&rows, `SELECT id, text, "sortOrder" FROM "QuotationTerm" ORDER BY "sortOrder", "updatedAt"`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الشروط")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/quotation-terms
func (h *QuotationTermsHandler) List(w http.ResponseWriter, r *http.Request) { h.list(w) }

// PUT /api/quotation-terms — يحفظ القائمة كاملة بترتيبها (إضافة/تعديل/حذف/ترتيب بخطوة وحدة).
func (h *QuotationTermsHandler) Save(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Terms []string `json:"terms"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	clean := []string{}
	for _, t := range body.Terms {
		if t = strings.TrimSpace(t); t != "" {
			if len([]rune(t)) > 500 {
				WriteError(w, http.StatusBadRequest, "الشرط طويل هواية — أقصى شي 500 حرف")
				return
			}
			clean = append(clean, t)
		}
	}
	if len(clean) > 30 {
		WriteError(w, http.StatusBadRequest, "أقصى شي 30 شرط")
		return
	}
	me := middleware.EmployeeIDFromContext(r)
	tx, err := h.db.Beginx()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM "QuotationTerm"`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	for i, t := range clean {
		if _, err := tx.Exec(`INSERT INTO "QuotationTerm" (text, "sortOrder", "updatedById") VALUES ($1, $2, NULLIF($3,''))`, t, i+1, me); err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	h.list(w)
}
