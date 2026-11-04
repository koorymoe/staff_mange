package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// MeetingRequestRepository «تذكيراتي»: طلبات مقابلة المدير وجلب الضيوف.
type MeetingRequestRepository struct{ db *sqlx.DB }

func NewMeetingRequestRepository(db *sqlx.DB) *MeetingRequestRepository {
	return &MeetingRequestRepository{db: db}
}

var ErrMeetingRequestDone = errors.New("هذا الطلب انحسم من قبل")

type MeetingRequest struct {
	ID          string     `db:"id" json:"id"`
	EmployeeID  string     `db:"employeeId" json:"-"`
	Who         string     `db:"who" json:"who,omitempty"`
	Kind        string     `db:"kind" json:"kind"`
	Reason      string     `db:"reason" json:"reason"`
	GuestName   *string    `db:"guestName" json:"guestName"`
	PreferredAt *time.Time `db:"preferredAt" json:"preferredAt"`
	Status      string     `db:"status" json:"status"`
	ScheduledAt *time.Time `db:"scheduledAt" json:"scheduledAt"`
	Note        *string    `db:"note" json:"note"`
	CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
}

const meetingCols = `m.id, m."employeeId", COALESCE(e.name, '') AS who, m.kind, m.reason, m."guestName",
	m."preferredAt", m.status, m."scheduledAt", m.note, m."createdAt"`

func (r *MeetingRequestRepository) Create(empID, kind, reason string, guest *string, preferred *time.Time) (*MeetingRequest, error) {
	id := uuid.NewString()
	if _, err := r.db.Exec(`INSERT INTO "MeetingRequest" (id, "employeeId", kind, reason, "guestName", "preferredAt")
		VALUES ($1, $2, $3, $4, $5, $6)`, id, empID, kind, reason, guest, preferred); err != nil {
		return nil, err
	}
	return r.Get(id)
}

func (r *MeetingRequestRepository) Get(id string) (*MeetingRequest, error) {
	var m MeetingRequest
	err := r.db.Get(&m, `SELECT `+meetingCols+` FROM "MeetingRequest" m
		LEFT JOIN "Employee" e ON e.id = m."employeeId" WHERE m.id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Mine طلبات الموظف نفسه (آخر ٣٠).
func (r *MeetingRequestRepository) Mine(empID string) ([]MeetingRequest, error) {
	rows := []MeetingRequest{}
	err := r.db.Select(&rows, `SELECT `+meetingCols+` FROM "MeetingRequest" m
		LEFT JOIN "Employee" e ON e.id = m."employeeId"
		WHERE m."employeeId" = $1 ORDER BY m."createdAt" DESC LIMIT 30`, empID)
	return rows, err
}

// Pending الطلبات الي تنتظر قرار — الأقدم أول.
func (r *MeetingRequestRepository) Pending() ([]MeetingRequest, error) {
	rows := []MeetingRequest{}
	err := r.db.Select(&rows, `SELECT `+meetingCols+` FROM "MeetingRequest" m
		LEFT JOIN "Employee" e ON e.id = m."employeeId"
		WHERE m.status = 'PENDING' ORDER BY m."createdAt" LIMIT 100`)
	return rows, err
}

// OpenCount طلبات الموظف المفتوحة — حد ضد الإغراق.
func (r *MeetingRequestRepository) OpenCount(empID string) (int, error) {
	var n int
	err := r.db.Get(&n, `SELECT count(*) FROM "MeetingRequest" WHERE "employeeId" = $1 AND status = 'PENDING'`, empID)
	return n, err
}

// Decide يحسم الطلب مرة وحدة: SCHEDULED بموعد، أو DECLINED بسبب.
func (r *MeetingRequestRepository) Decide(id, byID, status string, at *time.Time, note string) error {
	res, err := r.db.Exec(`UPDATE "MeetingRequest" SET status = $2, "scheduledAt" = $3, note = NULLIF($4, ''),
		"decidedById" = $5, "decidedAt" = now() WHERE id = $1 AND status = 'PENDING'`, id, status, at, note, byID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrMeetingRequestDone
	}
	return nil
}
