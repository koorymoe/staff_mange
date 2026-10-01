package repository

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type MatrixGuideRepository struct{ db *sqlx.DB }

func NewMatrixGuideRepository(db *sqlx.DB) *MatrixGuideRepository { return &MatrixGuideRepository{db: db} }

func (r *MatrixGuideRepository) List() ([]model.MatrixGuideRule, error) {
	rows := []model.MatrixGuideRule{}
	err := r.db.Select(&rows, `SELECT * FROM "MatrixGuideRule" ORDER BY route, priority DESC, "createdAt"`)
	return rows, err
}

func (r *MatrixGuideRepository) Enabled() ([]model.MatrixGuideRule, error) {
	rows := []model.MatrixGuideRule{}
	err := r.db.Select(&rows, `SELECT * FROM "MatrixGuideRule" WHERE enabled ORDER BY priority DESC, length(route) DESC`)
	return rows, err
}

func cleanGuideRule(in model.MatrixGuideRule) (model.MatrixGuideRule, error) {
	in.Route = strings.TrimSpace(in.Route)
	in.Text = strings.TrimSpace(in.Text)
	in.Match = strings.TrimSpace(in.Match)
	in.Groups = strings.TrimSpace(in.Groups)
	if in.Route == "" || !strings.HasPrefix(in.Route, "/") {
		return in, errors.New("المسار لازم يبدي بـ/ (مثلاً /coordinator)")
	}
	if in.Text == "" {
		return in, errors.New("اكتب نص التوجيه")
	}
	if len([]rune(in.Text)) > 400 {
		return in, errors.New("التوجيه طويل — خليه جملة أو جملتين")
	}
	return in, nil
}

func (r *MatrixGuideRepository) Create(in model.MatrixGuideRule) (*model.MatrixGuideRule, error) {
	in, err := cleanGuideRule(in)
	if err != nil {
		return nil, err
	}
	var out model.MatrixGuideRule
	err = r.db.Get(&out, `INSERT INTO "MatrixGuideRule" (id, route, match, groups, text, "onlyIfPending", priority, enabled)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *`,
		uuid.NewString(), in.Route, in.Match, in.Groups, in.Text, in.OnlyIfPending, in.Priority, in.Enabled)
	return &out, err
}

func (r *MatrixGuideRepository) Update(id string, in model.MatrixGuideRule) (*model.MatrixGuideRule, error) {
	in, err := cleanGuideRule(in)
	if err != nil {
		return nil, err
	}
	var out model.MatrixGuideRule
	err = r.db.Get(&out, `UPDATE "MatrixGuideRule" SET route=$2, match=$3, groups=$4, text=$5, "onlyIfPending"=$6, priority=$7, enabled=$8
		WHERE id=$1 RETURNING *`, id, in.Route, in.Match, in.Groups, in.Text, in.OnlyIfPending, in.Priority, in.Enabled)
	if err != nil {
		return nil, errors.New("التعليمة مو موجودة")
	}
	return &out, nil
}

func (r *MatrixGuideRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "MatrixGuideRule" WHERE id=$1`, id)
	return err
}
