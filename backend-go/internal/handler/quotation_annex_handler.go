package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// QuotationAnnexHandler (دازاي 11-06): قوالب الملحق الفني لـ«عرض المشاريع» — «يتغيّر حسب
// نوع المشروع». القالب: اسم وأجزاء، وكل جزء عنوان وبنود (عنوان + نص).
type QuotationAnnexHandler struct{ db *sqlx.DB }

func NewQuotationAnnexHandler(db *sqlx.DB) *QuotationAnnexHandler { return &QuotationAnnexHandler{db: db} }

type annexItem struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}
type annexPart struct {
	Title string      `json:"title"`
	Items []annexItem `json:"items"`
}
type annexTemplate struct {
	ID    string          `db:"id" json:"id"`
	Name  string          `db:"name" json:"name"`
	Parts json.RawMessage `db:"parts" json:"parts"`
}

// GET /api/quotation-annex-templates
func (h *QuotationAnnexHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []annexTemplate{}
	if err := h.db.Select(&rows, `SELECT id, name, parts FROM "QuotationAnnexTemplate" ORDER BY "sortOrder", name`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// validAnnex ينظّف الأجزاء: يشيل الفاضي، ولازم بند واحد على الأقل.
func validAnnex(parts []annexPart) ([]annexPart, bool) {
	out := []annexPart{}
	n := 0
	for _, p := range parts {
		p.Title = strings.TrimSpace(p.Title)
		items := []annexItem{}
		for _, it := range p.Items {
			it.Title, it.Text = strings.TrimSpace(it.Title), strings.TrimSpace(it.Text)
			if it.Title != "" || it.Text != "" {
				items = append(items, it)
			}
		}
		if len(items) > 0 {
			p.Items = items
			out = append(out, p)
			n += len(items)
		}
	}
	return out, n > 0
}

// PUT /api/quotation-annex-templates — {id?, name, parts}: إضافة أو تعديل.
func (h *QuotationAnnexHandler) Save(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ID    string      `json:"id"`
		Name  string      `json:"name"`
		Parts []annexPart `json:"parts"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	parts, ok := validAnnex(in.Parts)
	if len([]rune(in.Name)) < 3 || !ok {
		WriteError(w, http.StatusBadRequest, "اكتب اسم القالب وبند واحد على الأقل")
		return
	}
	b, _ := json.Marshal(parts)
	emp := middleware.EmployeeIDFromContext(r)
	var id string
	var err error
	if in.ID == "" {
		err = h.db.Get(&id, `INSERT INTO "QuotationAnnexTemplate" (name, parts, "updatedById") VALUES ($1, $2::jsonb, NULLIF($3, '')) RETURNING id`, in.Name, string(b), emp)
	} else {
		err = h.db.Get(&id, `UPDATE "QuotationAnnexTemplate" SET name = $2, parts = $3::jsonb, "updatedById" = NULLIF($4, ''), "updatedAt" = now() WHERE id = $1 RETURNING id`, in.ID, in.Name, string(b), emp)
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ (يمكن الاسم مكرر)")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"id": id})
}
