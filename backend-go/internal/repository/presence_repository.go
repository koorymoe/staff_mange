package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type PresenceRepository struct{ db *sqlx.DB }

func NewPresenceRepository(db *sqlx.DB) *PresenceRepository { return &PresenceRepository{db: db} }

func (r *PresenceRepository) Touch(employeeID string) {
	_, _ = r.db.Exec(`INSERT INTO "EmployeePresence" ("employeeId", "lastSeenAt") VALUES ($1, now())
		ON CONFLICT ("employeeId") DO UPDATE SET "lastSeenAt" = now()`, employeeID)
}

type PresenceRow struct {
	ID         string    `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Role       string    `db:"role" json:"role"`
	PhotoURL   *string   `db:"photoUrl" json:"photoUrl"`
	LastSeenAt time.Time `db:"lastSeenAt" json:"lastSeenAt"`
}

// Recent الموظفين الي ظهروا بآخر minutes دقيقة — الأحدث أول.
func (r *PresenceRepository) Recent(minutes int) ([]PresenceRow, error) {
	rows := []PresenceRow{}
	err := r.db.Select(&rows, `
		SELECT e.id, e.name, e.role::text AS role, e."photoUrl", p."lastSeenAt"
		FROM "EmployeePresence" p JOIN "Employee" e ON e.id = p."employeeId"
		WHERE p."lastSeenAt" > now() - make_interval(mins => $1)
		ORDER BY p."lastSeenAt" DESC`, minutes)
	return rows, err
}
