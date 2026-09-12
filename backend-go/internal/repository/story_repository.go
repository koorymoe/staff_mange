package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// StoryRepository طابور قصص الكيان.
type StoryRepository struct{ db *sqlx.DB }

func NewStoryRepository(db *sqlx.DB) *StoryRepository { return &StoryRepository{db: db} }

// openStatuses الحالات الي تعتبر القصة فيها «لسه ما انتهت».
const openStatuses = `('QUEUED','DELIVERED','PLAYING')`

// Enqueue ينشئ قصة جديدة.
//
// ⚠️ **الـidempotency من الفهرس مو من الكود**: `ON CONFLICT DO NOTHING`
// على `(eventId, recipientEmployeeId)` — نداء مكرر أو إعادة تشغيل
// **ما تنشئ قصة ثانية**، ويرجّع `created=false` بلا خطأ. نفس نمط
// `discipline_event_unique_penalty` الموجود بالمشروع.
//
// ⚠️ **والتجميع شغلة ثانية**: لو انكسر فهرس `groupKey` (يعني عند
// الموظف قصة مفتوحة بنفس المجموعة)، ما ننشئ صفاً ثانياً — **ندمج**
// بـ`MergeIntoGroup`. الفهرس الأول يمنع **التكرار**، والثاني يمنع
// **الازدحام**.
func (r *StoryRepository) Enqueue(s model.StoryInstance) (created bool, err error) {
	if s.EventID == "" || s.RecipientRef == "" {
		return false, errors.New("eventId و recipientRef إجباريان")
	}
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if len(s.Payload) == 0 {
		s.Payload = json.RawMessage(`{}`)
	}
	res, err := r.db.Exec(`
		INSERT INTO "StoryInstance"
			(id, "eventId", "eventKind", "storyType", version, "senderEmployeeId", "senderName",
			 "recipientEmployeeId", "recipientRef", "recipientName",
			 status, priority, physical, "groupKey", "currentStep", payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'QUEUED',$11,$12,$13,0,$14)
		ON CONFLICT DO NOTHING
	`, s.ID, s.EventID, s.EventKind, s.StoryType, 1, s.SenderEmployeeID, s.SenderName,
		s.RecipientEmployeeID, s.RecipientRef, s.RecipientName,
		s.Priority, s.Physical, s.GroupKey, s.Payload)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// MergeIntoGroup يضم حدثاً جديداً لقصة مفتوحة بنفس المجموعة.
//
// ⚠️ **ثلاث أوراق ناقصة بنفس الحجز = مشهد واحد يذكر الثلاثة**، مو
// ثلاث ركضات ورا بعض. نزيد العدّاد بالـpayload ونضيف السطر لقائمة
// البنود — والقصة **ما تنعاد للطابور من أولها** لو چانت تنعرض.
func (r *StoryRepository) MergeIntoGroup(recipientID, groupKey, line string) (merged bool, err error) {
	res, err := r.db.Exec(`
		UPDATE "StoryInstance"
		SET payload = jsonb_set(
		        jsonb_set(payload, '{mergedCount}',
		                  to_jsonb(COALESCE((payload->>'mergedCount')::int, 1) + 1)),
		        '{lines}',
		        COALESCE(payload->'lines', '[]'::jsonb) || to_jsonb($3::text))
		WHERE "recipientRef" = $1 AND "groupKey" = $2
		  AND status IN `+openStatuses, recipientID, groupKey, line)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// CountPhysicalToday يعدّ المشاهد الجسدية الي استلمها الموظف اليوم
// **بتوقيت بغداد** — نفس `baghdad_date` المستعملة بكل النظام، حتى
// السقف اليومي ما ينزل بمنتصف ليل غلط.
func (r *StoryRepository) CountPhysicalToday(recipientID string) (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "StoryInstance"
		WHERE "recipientRef" = $1 AND physical = true
		  AND baghdad_date("createdAt") = baghdad_today()`, recipientID)
	return n, err
}

// NextForEmployee يرجّع القصة الي دورها الآن — **وحدة بس**.
//
// ⚠️ **«ماكو قصة» مو خطأ**: يرجّع `(nil, nil)`. الأغلبية الساحقة من
// الاستطلاعات تجي من موظف نظيف، ولو عددناها خطأ يصير كل موظف يشوف
// ٥٠٠ كل دقيقة — لگاها الفحص الحي فعلاً.
//
// ⚠️ **قصة جسدية وحدة تلعب بالوقت** (شرط (م) نفسه: «لا تعمل قصتان
// جسديتان معاً»). الترتيب: الأعلى أولوية ثم الأقدم — يعني العقوبة
// تسبق الاحتفال حتى لو إجت بعده.
func (r *StoryRepository) NextForEmployee(recipientID string) (*model.StoryInstance, error) {
	var s model.StoryInstance
	err := r.db.Get(&s, `
		SELECT * FROM "StoryInstance"
		WHERE "recipientRef" = $1 AND status IN `+openStatuses+`
		  AND ("expiresAt" IS NULL OR "expiresAt" > now())
		ORDER BY priority DESC, "createdAt" ASC
		LIMIT 1`, recipientID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Claim يحجز القصة لنافذة وحدة — **عملية ذرّية**.
//
// ⚠️⚠️ **بلا هذا، نافذتان لنفس الموظف تشغّلان نفس المشهد**: نفس
// العقوبة تنعرض مرتين، والإقرار ينسجّل من وحدة والثانية تبقى عالقة
// تعرض شي انبتّ فيه. `RETURNING` يخلّي الحجز والفحص **عملية وحدة**
// بالقاعدة — نافذة تربح وحدة بس، والثانية تاخذ `false` وتسكت.
//
// ⚠️ **والحجز ما يعيد قصة تشتغل**: الشرط `status <> 'PLAYING'` يمنع
// نافذة ثانية من خطفها وسط المشهد.
func (r *StoryRepository) Claim(id, recipientRef string) (claimed bool, err error) {
	var got string
	err = r.db.Get(&got, `
		UPDATE "StoryInstance"
		SET status = 'PLAYING', "deliveredAt" = COALESCE("deliveredAt", now())
		WHERE id = $1 AND "recipientRef" = $2
		  AND status IN ('QUEUED','DELIVERED')
		RETURNING id`, id, recipientRef)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // نافذة ثانية سبقتنا، أو انبتّ فيها — مو خطأ
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// PendingCount عدد الي ينتظر الموظف — يخدم عدّاد الصندوق.
func (r *StoryRepository) PendingCount(recipientID string) (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "StoryInstance"
		WHERE "recipientRef" = $1 AND status IN `+openStatuses, recipientID)
	return n, err
}

// ListForEmployee صندوق الموظف — الأحدث أول.
func (r *StoryRepository) ListForEmployee(recipientID string, limit int) ([]model.StoryInstance, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows := []model.StoryInstance{}
	err := r.db.Select(&rows, `
		SELECT * FROM "StoryInstance"
		WHERE "recipientRef" = $1
		ORDER BY "createdAt" DESC LIMIT $2`, recipientID, limit)
	return rows, err
}

// storyStamps يربط كل مرحلة بعمودها الزمني.
//
// ⚠️ **أربع حقائق مستقلة**: «وصل» غير «انعرض» غير «انفتح» غير
// «أقرّ». و**الخروج من شاشة المراقب ليس قراءة** — شرط (م) الحرفي،
// ولذلك ماكو مرحلة تنكتب من طرف المرسِل إطلاقاً.
var storyStamps = map[string]string{
	model.StoryStatusDelivered:    "deliveredAt",
	model.StoryStatusSeen:         "seenAt",
	model.StoryStatusOpened:       "openedAt",
	model.StoryStatusAcknowledged: "acknowledgedAt",
}

// storyRank ترتيب المراحل — المرحلة ما ترجع للورا.
var storyRank = map[string]int{
	model.StoryStatusQueued: 0, model.StoryStatusDelivered: 1,
	model.StoryStatusPlaying: 2, model.StoryStatusSeen: 3,
	model.StoryStatusOpened: 4, model.StoryStatusAcknowledged: 5,
}

// Advance ينقل القصة لمرحلة أبعد ويختم طابعها الزمني.
//
// ⚠️ **`recipientID` بالشرط إجباري**: الموظف ما يقدر يعلّم قصة غيره
// مهما بدّل المعرّف بالرابط — نفس درس `/training/materials/mine`.
//
// ⚠️ **والمرحلة ما ترجع للورا**: `WHERE` يفرض إن الجديدة أبعد من
// الحالية. بلا هذا الشرط، استطلاع متأخر يقدر يرجّع «أقرّ» لـ«وصل»،
// فنخسر حقيقة مسجَّلة.
func (r *StoryRepository) Advance(id, recipientID, status string, step int) error {
	col, ok := storyStamps[status]
	if !ok && status != model.StoryStatusPlaying {
		return errors.New("مرحلة غير معروفة: " + status)
	}
	set := `status = $3, "currentStep" = GREATEST("currentStep", $4)`
	if ok {
		set += `, "` + col + `" = COALESCE("` + col + `", now())`
	}
	res, err := r.db.Exec(`
		UPDATE "StoryInstance" SET `+set+`
		WHERE id = $1 AND "recipientRef" = $2
		  AND $5 > CASE status
		        WHEN 'QUEUED' THEN 0 WHEN 'DELIVERED' THEN 1 WHEN 'PLAYING' THEN 2
		        WHEN 'SEEN' THEN 3 WHEN 'OPENED' THEN 4 WHEN 'ACKNOWLEDGED' THEN 5
		        ELSE 99 END`,
		id, recipientID, status, step, storyRank[status])
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		// مو خطأ: يعني القصة مو مالته، أو وصلت لمرحلة أبعد أصلاً.
		return nil
	}
	return nil
}

// ExpireStale يختم القصص الي بقت معلّقة أكثر من المدة.
//
// ⚠️ **ما تنمحى**: تصير `FAILED` وتبقى بالسجل. قصة انمحت يعني ماكو
// جواب على «ليش ما وصلته الرسالة؟».
func (r *StoryRepository) ExpireStale(olderThan time.Duration) (int64, error) {
	res, err := r.db.Exec(`
		UPDATE "StoryInstance" SET status = 'FAILED'
		WHERE status IN `+openStatuses+` AND "createdAt" < now() - $1::interval`,
		olderThan.String())
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}
