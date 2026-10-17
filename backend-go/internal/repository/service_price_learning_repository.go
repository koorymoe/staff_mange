package repository

import (
	"errors"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ═══ عيّنات أسعار الخدمات واقتراح المعدّل ═══
//
// كل فاتورة يدوية مربوطة بخدمة تنولد منها **عيّنة**. وبعد
// `ServicePriceMinSamples` عيّنة، النظام يحسب المعدّل ويسجّله
// **اقتراحاً** — مو سعراً.
//
// ⚠️ الاقتراح ما ينكتب على جدول الأسعار أبداً بهذا الملف: الكتابة
// هناك قرار مالك، وتصير بمسار الاعتماد وحده.

type ServicePriceLearningRepository struct {
	db *sqlx.DB
}

func NewServicePriceLearningRepository(db *sqlx.DB) *ServicePriceLearningRepository {
	return &ServicePriceLearningRepository{db: db}
}

// AddSample يسجّل عيّنة سعر. تنتنادى من مسار الفاتورة اليدوية.
//
// ⚠️ ما ترجّع خطأ يوقف الفاتورة: العيّنة **تحسين**، والفاتورة شغل
// الموظف. لو فشل تسجيل العيّنة، الفاتورة تنحفظ والعيّنة تضيع —
// والعكس (نوقف الفاتورة لأن العيّنة فشلت) يخلي ميزة تحليلية تمنع
// موظفاً من شغله.
func (r *ServicePriceLearningRepository) AddSample(serviceID, invoiceID, takenByID string, amount float64) error {
	_, err := r.db.Exec(`
		INSERT INTO "ServicePriceSample" (id, "serviceId", "invoiceId", amount, "takenById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)`,
		serviceID, nullIfEmpty(invoiceID), amount, nullIfEmpty(takenByID))
	return err
}

// Stats خلاصة عيّنات كل خدمة إلها عيّنات — مرتّبة بالأكثر عيّنات.
//
// ⚠️ المستثنى ما يدخل بالحساب، والعدد هو عدد الي **دخل** فعلاً —
// حتى ما يطلع «معدّل من ٧ عيّنات» وإحداها مستثناة.
func (r *ServicePriceLearningRepository) Stats() ([]model.ServicePriceStats, error) {
	rows := []model.ServicePriceStats{}
	err := r.db.Select(&rows, `
		SELECT s.id AS "serviceId", s.name AS "serviceName",
		       COUNT(*)          AS "sampleCount",
		       AVG(p.amount)     AS "avgAmount",
		       MIN(p.amount)     AS "minAmount",
		       MAX(p.amount)     AS "maxAmount"
		FROM "ServicePriceSample" p
		JOIN "Service" s ON s.id = p."serviceId"
		WHERE NOT p.excluded
		GROUP BY s.id, s.name
		ORDER BY COUNT(*) DESC, s.name`)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Remaining = model.ServicePriceMinSamples - rows[i].SampleCount
		if rows[i].Remaining < 0 {
			rows[i].Remaining = 0
		}
	}
	return rows, nil
}

// ProposeReady يبني اقتراحاً لكل خدمة وصلت الحد وما انطلع إلها
// اقتراح **أبداً** من قبل.
//
// 🔴 **اقتراح واحد لكل خدمة، مرة وحدة بالعمر** — قرار (ع) نصاً:
// «ما يقترح مرة ثانية أبداً» بعد ما يعتمد السعر. فالشرط مو «ماكو
// اقتراح مفتوح»، هو **ماكو ولا صف اقتراح** لهاي الخدمة.
//
// ⚠️ والمرفوض terminal هم: لو رجعنا نقترح لخدمة رفضها المالك، نرجع
// نزعجه بنفس الرقم تقريباً. والعيّنات تستمر تتجمّع وتبان بشاشة
// العيّنات، فالمعدّل الحيّ يبقى مقروءاً وهو يبدّل السعر بإيده وقت
// ما يريد — المعلومة محفوظة بلا إزعاج.
//
// ⚠️ و**ON CONFLICT DO NOTHING** يبقى حارساً ثانياً: لو انسوّت
// فاتورتان بنفس اللحظة، الفهرس الفريد الجزئي يمنع صفين لنفس الخدمة
// حتى لو مرّ الاثنان من فحص NOT EXISTS.
func (r *ServicePriceLearningRepository) ProposeReady() (int, error) {
	res, err := r.db.Exec(`
		INSERT INTO "ServicePriceSuggestion"
			(id, "serviceId", "avgAmount", "sampleCount", "minAmount", "maxAmount", source, status)
		SELECT gen_random_uuid()::text, p."serviceId",
		       AVG(p.amount), COUNT(*), MIN(p.amount), MAX(p.amount),
		       'MANUAL_AVERAGE', 'PROPOSED'
		FROM "ServicePriceSample" p
		WHERE NOT p.excluded
		  AND NOT EXISTS (
			SELECT 1 FROM "ServicePriceSuggestion" g
			WHERE g."serviceId" = p."serviceId"
		  )
		GROUP BY p."serviceId"
		HAVING COUNT(*) >= $1
		ON CONFLICT DO NOTHING`, model.ServicePriceMinSamples)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// Suggestions الاقتراحات — status فاضي يعني المفتوحة بس (طابور العمل).
func (r *ServicePriceLearningRepository) Suggestions(status string) ([]model.ServicePriceSuggestion, error) {
	rows := []model.ServicePriceSuggestion{}
	cond := `g.status IN ('PROPOSED', 'REVIEWED')`
	args := []any{}
	if status != "" {
		cond = `g.status = $1`
		args = append(args, status)
	}
	err := r.db.Select(&rows, `
		SELECT g.id, g."serviceId", s.name AS "serviceName", g."avgAmount", g."sampleCount",
		       g."minAmount", g."maxAmount", g.source, g.status,
		       g."reviewedById", rv.name AS "reviewedName", g."reviewedAt", g."reviewNote",
		       g."decidedById", dc.name AS "decidedName", g."decidedAt", g."createdAt"
		FROM "ServicePriceSuggestion" g
		LEFT JOIN "Service"  s  ON s.id  = g."serviceId"
		LEFT JOIN "Employee" rv ON rv.id = g."reviewedById"
		LEFT JOIN "Employee" dc ON dc.id = g."decidedById"
		WHERE `+cond+`
		ORDER BY g."createdAt" DESC`, args...)
	return rows, err
}

// MarkReviewed المحاسب راجع الاقتراح وصادق إنه معقول — خطوة قبل
// المالك، مثل ما طلب (ع): «يقترح للمالك والمحاسب يراجع قبلك».
func (r *ServicePriceLearningRepository) MarkReviewed(id, byEmployeeID, note string) error {
	res, err := r.db.Exec(`
		UPDATE "ServicePriceSuggestion"
		SET status = 'REVIEWED', "reviewedById" = $2, "reviewedAt" = now(),
		    "reviewNote" = NULLIF($3, '')
		WHERE id = $1 AND status = 'PROPOSED'`, id, nullIfEmpty(byEmployeeID), note)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("هذا الاقتراح مو بانتظار مراجعة المحاسب")
	}
	return nil
}

// Decide المالك يعتمد أو يرفض.
//
// ⚠️ الاعتماد **لازم يمر بمراجعة المحاسب أولاً** (status = REVIEWED)
// — هذا ترتيب (ع) نصاً. أما الرفض فيصير بأي حالة مفتوحة: المالك ما
// ينتظر مراجعة حتى يرفض رقماً يشوفه غلطاً.
func (r *ServicePriceLearningRepository) Decide(id, byEmployeeID string, approve bool) error {
	target := model.PriceSuggestionApproved
	cond := `status = 'REVIEWED'`
	if !approve {
		target = model.PriceSuggestionRejected
		cond = `status IN ('PROPOSED', 'REVIEWED')`
	}
	res, err := r.db.Exec(`
		UPDATE "ServicePriceSuggestion"
		SET status = $3, "decidedById" = $2, "decidedAt" = now()
		WHERE id = $1 AND `+cond, id, nullIfEmpty(byEmployeeID), target)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if approve {
			return errors.New("لازم المحاسب يراجعه قبل الاعتماد")
		}
		return errors.New("هذا الاقتراح انبتّ بيه من قبل")
	}
	return nil
}
