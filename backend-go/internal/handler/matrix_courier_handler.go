package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ رسول ماتركس — قرار (ع) 10-13 ═══
//
// المالك أو المدير يوجّه شغل لموظف (مهمة إضافية، موافقة إجازة، مشروع،
// حجز) → عين الشارنغان تتحول مانغكيو وتغمض وتختفي من شاشته، تطلع يم عين
// الموظف وتسلّمه الرسالة، وترجع. السطر يبقى حتى الموظف يرجع يقرا الرسالة
// من عينه، والبند الي توجّه له الشغل يطلع عليه رقم لحد ما يفتحه.
//
// ⚠️ الرسول ما يغيّر الإجراء نفسه: يلتف حول الهاندلر، ولمن يرجع 2xx
// ويكون الفاعل مالك/مدير، يسجّل سطر. أي فشل بالتسجيل ينكتب باللوك بس.

type MatrixCourierHandler struct{ db *sqlx.DB }

func NewMatrixCourierHandler(db *sqlx.DB) *MatrixCourierHandler {
	return &MatrixCourierHandler{db: db}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) { s.status = code; s.ResponseWriter.WriteHeader(code) }

// Wrap يلف هاندلر إجراء من نوع kind (extra_task | leave | project | booking).
func (h *MatrixCourierHandler) Wrap(kind string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if role := middleware.RoleFromContext(r); role != "ADMIN" && role != "OWNER" {
				next.ServeHTTP(w, r)
				return
			}
			raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			r.Body = io.NopCloser(bytes.NewReader(raw))
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if rec.status < 200 || rec.status >= 300 {
				return
			}
			h.record(kind, middleware.EmployeeIDFromContext(r), r.PathValue("id"), raw)
		})
	}
}

func (h *MatrixCourierHandler) record(kind, senderID, pathID string, raw []byte) {
	var body struct {
		EmployeeID   string `json:"employeeId"`
		AssignedToID string `json:"assignedToId"`
		Title        string `json:"title"`
		Approve      bool   `json:"approve"`
		Note         string `json:"note"`
	}
	_ = json.Unmarshal(raw, &body)

	var recipient, message, link string
	switch kind {
	case "extra_task":
		recipient = body.AssignedToID
		message = "المدير وجّهلك مهمة إضافية: «" + body.Title + "»"
		link = "/my-work"
	case "leave":
		if !body.Approve {
			return
		}
		if err := h.db.Get(&recipient, `SELECT "employeeId" FROM "LeaveRequest" WHERE id = $1`, pathID); err != nil {
			return
		}
		message = "المدير وافق على طلب إجازتك ✅"
		link = "/attendance"
	case "project":
		recipient = body.EmployeeID
		var name string
		_ = h.db.Get(&name, `SELECT name FROM "Project" WHERE id = $1`, pathID)
		message = "المدير وجّهلك مشروع: «" + name + "»"
		link = "/my-projects"
	case "booking":
		recipient = body.EmployeeID
		var code string
		_ = h.db.Get(&code, `SELECT code FROM "Booking" WHERE id = $1`, pathID)
		message = "المدير وجّهلك الحجز " + code
		link = "/my-tasks"
	default:
		return
	}
	if body.Note != "" && kind != "extra_task" {
		message += " — " + body.Note
	}
	if recipient == "" || recipient == senderID {
		return
	}
	if _, err := h.db.Exec(`INSERT INTO "MatrixCourier" (id, kind, "senderId", "recipientId", message, link) VALUES ($1,$2,$3,$4,$5,$6)`,
		uuid.NewString(), kind, senderID, recipient, message, link); err != nil {
		log.Printf("matrix courier %s: %v", kind, err)
	}
}

type courierRow struct {
	ID            string     `db:"id" json:"id"`
	Kind          string     `db:"kind" json:"kind"`
	Message       string     `db:"message" json:"message"`
	Link          string     `db:"link" json:"link"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	DeliveredAt   *time.Time `db:"deliveredAt" json:"deliveredAt"`
	RecipientName string     `db:"recipientName" json:"recipientName,omitempty"`
}

// GET /api/matrix-courier/state — صادر (للمرسل، آخر دقيقة) + وارد ما وصل + أرقام البنود.
func (h *MatrixCourierHandler) State(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	out := struct {
		Now      time.Time      `json:"now"`
		Outgoing []courierRow   `json:"outgoing"`
		Incoming []courierRow   `json:"incoming"`
		Badges   map[string]int `json:"badges"`
	}{Now: time.Now(), Outgoing: []courierRow{}, Incoming: []courierRow{}, Badges: map[string]int{}}
	_ = h.db.Select(&out.Outgoing, `SELECT c.id, c.kind, c.message, c.link, c."createdAt", c."deliveredAt", e.name AS "recipientName"
		FROM "MatrixCourier" c JOIN "Employee" e ON e.id = c."recipientId"
		WHERE c."senderId" = $1 AND c."createdAt" > now() - interval '60 seconds' ORDER BY c."createdAt"`, me)
	_ = h.db.Select(&out.Incoming, `SELECT id, kind, message, link, "createdAt", "deliveredAt" FROM "MatrixCourier"
		WHERE "recipientId" = $1 AND "deliveredAt" IS NULL AND "createdAt" > now() - interval '3 days' ORDER BY "createdAt" LIMIT 5`, me)
	rows := []struct {
		Link string `db:"link"`
		N    int    `db:"n"`
	}{}
	_ = h.db.Select(&rows, `SELECT link, count(*) AS n FROM "MatrixCourier" WHERE "recipientId" = $1 AND "seenAt" IS NULL GROUP BY link`, me)
	for _, x := range rows {
		out.Badges[x.Link] = x.N
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/matrix-courier/inbox — آخر رسائل ماتركس الي وصلته، حتى يرجع يقراها.
func (h *MatrixCourierHandler) Inbox(w http.ResponseWriter, r *http.Request) {
	rows := []courierRow{}
	_ = h.db.Select(&rows, `SELECT id, kind, message, link, "createdAt", "deliveredAt" FROM "MatrixCourier"
		WHERE "recipientId" = $1 ORDER BY "createdAt" DESC LIMIT 20`, middleware.EmployeeIDFromContext(r))
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/matrix-courier/{id}/delivered — العين سلّمت الرسالة على شاشته.
func (h *MatrixCourierHandler) Delivered(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`UPDATE "MatrixCourier" SET "deliveredAt" = now() WHERE id = $1 AND "recipientId" = $2 AND "deliveredAt" IS NULL`,
		r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/matrix-courier/seen?link=/my-projects — فتح البند فينصفّر رقمه.
func (h *MatrixCourierHandler) Seen(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`UPDATE "MatrixCourier" SET "seenAt" = now(), "deliveredAt" = COALESCE("deliveredAt", now())
		WHERE "recipientId" = $1 AND link = $2 AND "seenAt" IS NULL`,
		middleware.EmployeeIDFromContext(r), r.URL.Query().Get("link"))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
