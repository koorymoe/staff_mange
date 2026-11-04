package repository

import (
	"database/sql"
	"errors"
	"strings"
)

// ═══ إرجاع مشروع مرفوض — طلب (ع) 10-19 ═══
// «اذا اكو مشروع مرفوض واريد ارجعه للمشاريع». يرجع للمرحلة الي چان بيها
// قبل الرفض (من ProjectStageLog)، والسبب ومنو رجّعه بـ"ProjectRestoreLog".
const projectRestoreFallbackStage = "1. اتصال بالزبون"

var ErrProjectNotRejected = errors.New("المشروع مو مرفوض — ماكو شي نرجّعه")

func (r *ProjectRepository) Restore(id, byID, reason string) (string, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	var stage string
	if err := tx.Get(&stage, `SELECT stage FROM "Project" WHERE id = $1 FOR UPDATE`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("المشروع مو موجود")
		}
		return "", err
	}
	if !strings.Contains(stage, "مرفوض") {
		return "", ErrProjectNotRejected
	}
	var back sql.NullString
	_ = tx.Get(&back, `SELECT "fromStage" FROM "ProjectStageLog"
		WHERE "projectId" = $1 AND "toStage" LIKE '%مرفوض%' AND "fromStage" IS NOT NULL AND "fromStage" NOT LIKE '%مرفوض%'
		ORDER BY "createdAt" DESC LIMIT 1`, id)
	to := projectRestoreFallbackStage
	if back.Valid && strings.TrimSpace(back.String) != "" {
		to = back.String
	}
	if _, err := tx.Exec(`UPDATE "Project" SET stage = $2, "updatedAt" = now() WHERE id = $1`, id, to); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`INSERT INTO "ProjectRestoreLog" ("projectId", "toStage", reason, "byId", "byName")
		VALUES ($1, $2, $3, NULLIF($4, ''), (SELECT name FROM "Employee" WHERE id = $4))`, id, to, reason, byID); err != nil {
		return "", err
	}
	return to, tx.Commit()
}

// ═══ حذف المشروع المرفوض نهائياً — (ع) 10-19 ═══
// المرفوض بس، والسبب إجباري. نسخة كاملة تنحفظ بـ"DeletedProject" قبل الحذف.
func (r *ProjectRepository) PurgeRejected(id, byID, reason string) (string, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	var p struct {
		Code  string `db:"code"`
		Name  string `db:"name"`
		Stage string `db:"stage"`
	}
	if err := tx.Get(&p, `SELECT COALESCE(code, '') AS code, COALESCE(name, '') AS name, stage FROM "Project" WHERE id = $1 FOR UPDATE`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("المشروع مو موجود")
		}
		return "", err
	}
	if !strings.Contains(p.Stage, "مرفوض") {
		return "", errors.New("الحذف النهائي للمشروع المرفوض بس — ارفضه أول")
	}
	if _, err := tx.Exec(`INSERT INTO "DeletedProject" ("projectId", code, name, snapshot, reason, "deletedById", "deletedByName")
		SELECT p.id, p.code, p.name, to_jsonb(p), $2, NULLIF($3, ''), (SELECT name FROM "Employee" WHERE id = $3)
		FROM "Project" p WHERE p.id = $1`, id, reason, byID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`DELETE FROM "Project" WHERE id = $1`, id); err != nil {
		return "", err
	}
	return p.Name, tx.Commit()
}
