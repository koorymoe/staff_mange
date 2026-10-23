package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ═══ 💬 محادثات وكيلي الواتساب وإجراءاتهم (دازاي 10-23 — docs/wa-api-contract.md) ═══

type WaRepository struct{ db *sqlx.DB }

func NewWaRepository(db *sqlx.DB) *WaRepository { return &WaRepository{db: db} }

type WaConversation struct {
	ID            string     `db:"id" json:"id"`
	Agent         string     `db:"agent" json:"agent"`
	CustomerID    *string    `db:"customerId" json:"customerId"`
	CustomerName  string     `db:"customerName" json:"customerName"`
	Phone         string     `db:"phone" json:"phone"`
	BookingID     *string    `db:"bookingId" json:"bookingId"`
	BookingCode   *string    `db:"bookingCode" json:"bookingCode"`
	FollowUpID    *string    `db:"followUpId" json:"followUpId"`
	Status        string     `db:"status" json:"status"`
	Sandbox       bool       `db:"sandbox" json:"sandbox"`
	HumanByID     *string    `db:"humanById" json:"humanById"`
	LastMessageAt time.Time  `db:"lastMessageAt" json:"lastMessageAt"`
	LastMessage   string     `db:"lastMessage" json:"lastMessage"`
	Unread        int        `db:"unread" json:"unread"`
	LastReadAt    *time.Time `db:"lastReadAt" json:"-"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
}

type WaMessage struct {
	ID             string    `db:"id" json:"id"`
	ConversationID string    `db:"conversationId" json:"-"`
	Direction      string    `db:"direction" json:"direction"`
	Author         string    `db:"author" json:"author"`
	AuthorID       *string   `db:"authorId" json:"-"`
	AuthorName     *string   `db:"authorName" json:"authorName,omitempty"`
	Text           string    `db:"text" json:"text"`
	CreatedAt      time.Time `db:"createdAt" json:"createdAt"`
}

type AiAgentAction struct {
	ID             string     `db:"id" json:"id"`
	Agent          string     `db:"agent" json:"agent"`
	AgentName      string     `db:"-" json:"agentName"`
	ConversationID *string    `db:"conversationId" json:"conversationId"`
	Kind           string     `db:"kind" json:"kind"`
	Summary        string     `db:"summary" json:"summary"`
	Reason         string     `db:"reason" json:"reason"`
	TargetType     *string    `db:"targetType" json:"targetType"`
	TargetID       *string    `db:"targetId" json:"targetId"`
	Payload        NullBytes  `db:"payload" json:"-"`
	Sandbox        bool       `db:"sandbox" json:"sandbox"`
	Undoable       bool       `db:"-" json:"undoable"`
	UndoneAt       *time.Time `db:"undoneAt" json:"undoneAt"`
	UndoneByName   *string    `db:"undoneByName" json:"undoneByName"`
	UndoReason     *string    `db:"undoReason" json:"undoReason,omitempty"`
	CreatedAt      time.Time  `db:"createdAt" json:"createdAt"`
}

// NullBytes JSONB اختياري.
type NullBytes []byte

func (n *NullBytes) Scan(v any) error {
	switch x := v.(type) {
	case nil:
		*n = nil
	case []byte:
		*n = append((*n)[:0], x...)
	case string:
		*n = []byte(x)
	default:
		return fmt.Errorf("NullBytes: %T", v)
	}
	return nil
}

const waConvSelect = `SELECT c.id, c.agent, c."customerId", COALESCE(cu.name, '') AS "customerName", c.phone,
	c."bookingId", b.code AS "bookingCode", c."followUpId", c.status, c.sandbox, c."humanById",
	c."lastMessageAt", c."lastReadAt", c."createdAt",
	COALESCE((SELECT left(m.text, 120) FROM "WaMessage" m WHERE m."conversationId" = c.id ORDER BY m."createdAt" DESC LIMIT 1), '') AS "lastMessage",
	(SELECT COUNT(*) FROM "WaMessage" m WHERE m."conversationId" = c.id AND m.direction = 'IN'
		AND m."createdAt" > COALESCE(c."lastReadAt", 'epoch'::timestamptz))::int AS unread
	FROM "WaConversation" c
	LEFT JOIN "Customer" cu ON cu.id = c."customerId"
	LEFT JOIN "Booking" b ON b.id = c."bookingId"`

var ErrWaNotFound = errors.New("المحادثة مو موجودة")

func (r *WaRepository) CreateConversation(agent string, customerID *string, phone string, bookingID, followUpID *string, sandbox bool) (*WaConversation, error) {
	id := uuid.NewString()
	if _, err := r.db.Exec(`INSERT INTO "WaConversation" (id, agent, "customerId", phone, "bookingId", "followUpId", sandbox)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, id, agent, customerID, phone, bookingID, followUpID, sandbox); err != nil {
		return nil, err
	}
	return r.Conversation(id)
}

func (r *WaRepository) Conversation(id string) (*WaConversation, error) {
	var c WaConversation
	if err := r.db.Get(&c, waConvSelect+` WHERE c.id = $1`, id); err != nil {
		return nil, ErrWaNotFound
	}
	return &c, nil
}

// OpenByPhone آخر محادثة حقيقية مفتوحة لهذا الرقم (الرسالة الواردة تروحلها).
func (r *WaRepository) OpenByPhone(phone string) (*WaConversation, error) {
	var c WaConversation
	if err := r.db.Get(&c, waConvSelect+` WHERE c.phone = $1 AND c.status <> 'CLOSED' AND NOT c.sandbox
		ORDER BY c."lastMessageAt" DESC LIMIT 1`, phone); err != nil {
		return nil, ErrWaNotFound
	}
	return &c, nil
}

func (r *WaRepository) Conversations(agent, status string) ([]WaConversation, error) {
	rows := []WaConversation{}
	err := r.db.Select(&rows, waConvSelect+` WHERE ($1 = '' OR c.agent = $1) AND ($2 = '' OR c.status = $2)
		ORDER BY c."lastMessageAt" DESC LIMIT 200`, agent, status)
	return rows, err
}

// AddMessage يحفظ رسالة. metaId المكرر (Meta تعيد الإرسال) يرجّع inserted=false.
func (r *WaRepository) AddMessage(convID, direction, author string, authorID *string, text string, metaID *string) (*WaMessage, bool, error) {
	var m WaMessage
	err := r.db.Get(&m, `INSERT INTO "WaMessage" (id, "conversationId", direction, author, "authorId", text, "metaId")
		VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT ("metaId") DO NOTHING
		RETURNING id, "conversationId", direction, author, "authorId", NULL::text AS "authorName", text, "createdAt"`,
		uuid.NewString(), convID, direction, author, authorID, text, metaID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, false, nil
		}
		return nil, false, err
	}
	_, _ = r.db.Exec(`UPDATE "WaConversation" SET "lastMessageAt" = now() WHERE id = $1`, convID)
	return &m, true, nil
}

func (r *WaRepository) Messages(convID string) ([]WaMessage, error) {
	rows := []WaMessage{}
	err := r.db.Select(&rows, `SELECT m.id, m."conversationId", m.direction, m.author, m."authorId", e.name AS "authorName", m.text, m."createdAt"
		FROM "WaMessage" m LEFT JOIN "Employee" e ON e.id = m."authorId"
		WHERE m."conversationId" = $1 ORDER BY m."createdAt", m.id LIMIT 500`, convID)
	return rows, err
}

func (r *WaRepository) SetStatus(id, status string, humanBy *string) error {
	_, err := r.db.Exec(`UPDATE "WaConversation" SET status = $2, "humanById" = $3 WHERE id = $1`, id, status, humanBy)
	return err
}

func (r *WaRepository) MarkRead(id string) {
	_, _ = r.db.Exec(`UPDATE "WaConversation" SET "lastReadAt" = now() WHERE id = $1`, id)
}

// NewToday محادثات حقيقية انفتحت اليوم لهذا الوكيل — حد WA_DAILY_NEW.
func (r *WaRepository) NewToday(agent string) int {
	var n int
	_ = r.db.Get(&n, `SELECT COUNT(*) FROM "WaConversation" WHERE agent = $1 AND NOT sandbox AND baghdad_date("createdAt") = baghdad_today()`, agent)
	return n
}

// ── الإجراءات ──

func (r *WaRepository) AddAction(a AiAgentAction, payload any) (*AiAgentAction, error) {
	var pl any
	if payload != nil {
		b, _ := json.Marshal(payload)
		pl = b
	}
	a.ID = uuid.NewString()
	_, err := r.db.Exec(`INSERT INTO "AiAgentAction" (id, agent, "conversationId", kind, summary, reason, "targetType", "targetId", payload, sandbox)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, a.ID, a.Agent, a.ConversationID, a.Kind, a.Summary, a.Reason, a.TargetType, a.TargetID, pl, a.Sandbox)
	if err != nil {
		return nil, err
	}
	return r.Action(a.ID)
}

const actionSelect = `SELECT a.id, a.agent, a."conversationId", a.kind, a.summary, a.reason, a."targetType", a."targetId", a.payload,
	a.sandbox, a."undoneAt", e.name AS "undoneByName", a."undoReason", a."createdAt"
	FROM "AiAgentAction" a LEFT JOIN "Employee" e ON e.id = a."undoneById"`

func (r *WaRepository) Action(id string) (*AiAgentAction, error) {
	var a AiAgentAction
	if err := r.db.Get(&a, actionSelect+` WHERE a.id = $1`, id); err != nil {
		return nil, errors.New("الإجراء مو موجود")
	}
	return &a, nil
}

func (r *WaRepository) Actions(targetID, convID string) ([]AiAgentAction, error) {
	rows := []AiAgentAction{}
	err := r.db.Select(&rows, actionSelect+` WHERE ($1 = '' OR a."targetId" = $1) AND ($2 = '' OR a."conversationId" = $2)
		ORDER BY a."createdAt" DESC LIMIT 200`, targetID, convID)
	return rows, err
}

func (r *WaRepository) MarkUndone(id, byID, reason string) error {
	res, err := r.db.Exec(`UPDATE "AiAgentAction" SET "undoneAt" = now(), "undoneById" = $2, "undoReason" = $3
		WHERE id = $1 AND "undoneAt" IS NULL`, id, byID, reason)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("الإجراء انلغى من قبل")
	}
	return nil
}

// ── سياق الوكيل وأفعاله على النظام ──

// WaBookingContext الي يشوفه الوكيل عن الحجز. ⚠️ بلا اسم الزبون ولا رقمه —
// لحد قرار الخصوصية المكتوب (AGENTS.md)؛ أسماء الكادر تنخفى بقناع ماتركس.
type WaBookingContext struct {
	BookingID   string     `db:"id" json:"-"`
	CustomerID  string     `db:"customerId" json:"-"`
	Phone       string     `db:"phone" json:"-"`
	Code        string     `db:"code" json:"code"`
	Status      string     `db:"status" json:"status"`
	Type        string     `db:"bookingType" json:"type"`
	Services    string     `db:"services" json:"services"`
	Notes       *string    `db:"notes" json:"notes"`
	ScheduledAt *time.Time `db:"scheduledAt" json:"scheduledAt"`
	CompletedAt *time.Time `db:"completedAt" json:"completedAt"`
	Leader      *string    `db:"leader" json:"leader"`
}

func (r *WaRepository) BookingContext(bookingID string) (*WaBookingContext, error) {
	var c WaBookingContext
	err := r.db.Get(&c, `SELECT b.id, b."customerId", COALESCE(cu.phone, '') AS phone, b.code, b.status::text AS status,
		b."bookingType"::text AS "bookingType",
		COALESCE((SELECT string_agg(s.name, '، ') FROM "Service" s WHERE s.id = b."serviceId"
			OR s.id IN (SELECT x."serviceId" FROM "BookingService" x WHERE x."bookingId" = b.id)), '') AS services,
		b.notes, b."scheduledAt", b."completedAt",
		(SELECT e.name FROM "Mission" mi JOIN "Employee" e ON e.id = mi."leaderId" WHERE mi."bookingId" = b.id ORDER BY mi."createdAt" DESC LIMIT 1) AS leader
		FROM "Booking" b JOIN "Customer" cu ON cu.id = b."customerId" WHERE b.id = $1`, bookingID)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// QualityCandidates متابعات PENDING لحجوزات خلصت آخر ٧ أيام، بلا محادثة — القديمة ما تنلمس.
func (r *WaRepository) QualityCandidates(limit int) ([]struct {
	FollowUpID string `db:"followUpId"`
	BookingID  string `db:"bookingId"`
}, error) {
	rows := []struct {
		FollowUpID string `db:"followUpId"`
		BookingID  string `db:"bookingId"`
	}{}
	err := r.db.Select(&rows, `SELECT q.id AS "followUpId", q."bookingId" FROM "QualityFollowUp" q
		JOIN "Booking" b ON b.id = q."bookingId" JOIN "Customer" cu ON cu.id = b."customerId"
		WHERE q.status = 'PENDING' AND b."completedAt" > now() - interval '7 days' AND COALESCE(cu.phone, '') <> ''
		  AND NOT EXISTS (SELECT 1 FROM "WaConversation" c WHERE c."followUpId" = q.id AND NOT c.sandbox)
		ORDER BY b."completedAt" LIMIT $1`, limit)
	return rows, err
}

// BookingCandidates حجوزات جديدة بانتظار التواصل (آخر يومين)، بلا محادثة.
func (r *WaRepository) BookingCandidates(limit int) ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `SELECT b.id FROM "Booking" b JOIN "Customer" cu ON cu.id = b."customerId"
		WHERE (`+bucketCondition("pending")+`) AND b."confirmationContactedAt" IS NULL
		  AND b."createdAt" > now() - interval '2 days' AND COALESCE(cu.phone, '') <> ''
		  AND NOT EXISTS (SELECT 1 FROM "WaConversation" c WHERE c."bookingId" = b.id AND c.agent = 'BOOKING' AND NOT c.sandbox)
		ORDER BY b."createdAt" LIMIT $1`, limit)
	return ids, err
}

// RecordContactByAgent «تم التواصل» من وكيل الجودة: الحالة والتقييمين والملاحظة.
func (r *WaRepository) RecordContactByAgent(followUpID, agent string, ok bool, customerRating, crewRating *int, notes string) error {
	status := "CONTACTED_OK"
	if !ok {
		status = "CONTACTED_ISSUE"
	}
	res, err := r.db.Exec(`UPDATE "QualityFollowUp" SET status = $2, "contactNotes" = NULLIF($3, ''),
		"contactedAt" = now(), "contactedByEmployeeId" = NULL, "contactedByAgent" = $4,
		"customerRating" = $5, "crewRating" = $6
		WHERE id = $1 AND status IN ('PENDING', 'RECONTACT')`, followUpID, status, notes, agent, customerRating, crewRating)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("المتابعة انسجّلت من قبل")
	}
	return nil
}

// AppendBookingNote يضيف سطر للملاحظات (تفاصيل من الزبون).
func (r *WaRepository) AppendBookingNote(bookingID, line string) error {
	_, err := r.db.Exec(`UPDATE "Booking" SET notes = CASE WHEN COALESCE(notes, '') = '' THEN $2 ELSE notes || E'\n' || $2 END WHERE id = $1`, bookingID, line)
	return err
}

// MarkContactedByAgent «تواصلت ويا الزبون» من وكيل الحجوزات.
func (r *WaRepository) MarkContactedByAgent(bookingID string) error {
	_, err := r.db.Exec(`UPDATE "Booking" SET "confirmationContactedAt" = now() WHERE id = $1 AND "confirmationContactedAt" IS NULL`, bookingID)
	return err
}
